package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Executor interface for different protocol executors
type Executor interface {
	Execute(ctx context.Context) error
}

// TestExecution represents a running test execution
type TestExecution struct {
	ID               string
	UseCaseID        string
	Method           string
	Path             string
	TargetAddress    string
	TargetProtocol   string
	ConnectionPooled bool
	SimulatedUsers   int
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	executor         Executor
	grpcExecutor     *GRPCExecutor // Keep reference for cleanup
}

// TestExecutionManager manages active test executions
type TestExecutionManager struct {
	mu         sync.RWMutex
	executions map[string]*TestExecution
}

// Global test execution manager
var globalTestManager = &TestExecutionManager{
	executions: make(map[string]*TestExecution),
}

// GetTestExecutionManager returns the global test execution manager
func GetTestExecutionManager() *TestExecutionManager {
	return globalTestManager
}

// StartTest starts a new test execution
func (m *TestExecutionManager) StartTest(testExecutionID, useCaseID, targetAddress, targetProtocol string, connectionPooled bool, simulatedUsers int, sni string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if test already exists
	if _, exists := m.executions[testExecutionID]; exists {
		return fmt.Errorf("test execution %s already running", testExecutionID)
	}

	// Get use case details
	useCase, err := getUseCaseByID(useCaseID)
	if err != nil {
		return fmt.Errorf("failed to get use case: %w", err)
	}

	// Create context for test execution
	ctx, cancel := context.WithCancel(context.Background())

	// Create test execution
	execution := &TestExecution{
		ID:               testExecutionID,
		UseCaseID:        useCaseID,
		Method:           useCase.Method,
		Path:             useCase.Path,
		TargetAddress:    targetAddress,
		TargetProtocol:   targetProtocol,
		ConnectionPooled: connectionPooled,
		SimulatedUsers:   simulatedUsers,
		ctx:              ctx,
		cancel:           cancel,
	}

	// Create executor based on protocol
	if targetProtocol == "grpc" {
		// Create gRPC executor
		grpcExecutor, err := NewGRPCExecutor(
			targetAddress,
			targetProtocol,
			connectionPooled,
			useCaseID,
			testExecutionID,
			sni,
		)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to create gRPC executor: %w", err)
		}
		execution.executor = grpcExecutor
		execution.grpcExecutor = grpcExecutor

		log.Printf("Created gRPC executor for target %s", targetAddress)
	} else {
		// Create HTTP executor
		execution.executor = NewHTTPExecutor(
			targetAddress,
			targetProtocol,
			connectionPooled,
			useCaseID,
			useCase.Method,
			useCase.Path,
			testExecutionID,
			sni,
		)

		log.Printf("Created HTTP executor for target %s", targetAddress)
	}

	// Store execution
	m.executions[testExecutionID] = execution

	// Update metrics
	AgentActiveTests.Inc()
	AgentActiveUsers.WithLabelValues(testExecutionID).Set(float64(simulatedUsers))

	// Start simulated users
	log.Printf("Starting test execution %s with %d users for use case %s (protocol: %s, connection pooling: %v)",
		testExecutionID, simulatedUsers, useCaseID, targetProtocol, connectionPooled)

	for i := 0; i < simulatedUsers; i++ {
		execution.wg.Add(1)
		go execution.runUser(i + 1)
	}

	log.Printf("Test execution %s started successfully", testExecutionID)
	return nil
}

// StopTest stops a running test execution
func (m *TestExecutionManager) StopTest(testExecutionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	execution, exists := m.executions[testExecutionID]
	if !exists {
		return fmt.Errorf("test execution %s not found", testExecutionID)
	}

	log.Printf("Stopping test execution %s", testExecutionID)

	// Cancel context to stop all users
	execution.cancel()

	// Wait for all users to finish (with timeout)
	done := make(chan struct{})
	go func() {
		execution.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("All users stopped for test execution %s", testExecutionID)
	case <-time.After(10 * time.Second):
		log.Printf("Timeout waiting for users to stop for test execution %s", testExecutionID)
	}

	// Clean up gRPC executor if exists
	if execution.grpcExecutor != nil {
		if err := execution.grpcExecutor.Close(); err != nil {
			log.Printf("Error closing gRPC executor: %v", err)
		}
	}

	// Remove execution
	delete(m.executions, testExecutionID)

	// Update metrics
	AgentActiveTests.Dec()
	AgentActiveUsers.DeleteLabelValues(testExecutionID)

	log.Printf("Test execution %s stopped", testExecutionID)
	return nil
}

// GetActiveTests returns the number of active test executions
func (m *TestExecutionManager) GetActiveTests() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.executions)
}

// runUser simulates a single user making requests in a loop
func (e *TestExecution) runUser(userID int) {
	defer e.wg.Done()

	log.Printf("User %d started for test execution %s", userID, e.ID)

	for {
		select {
		case <-e.ctx.Done():
			log.Printf("User %d stopping for test execution %s", userID, e.ID)
			return
		default:
			// Execute request
			err := e.executor.Execute(e.ctx)
			if err != nil {
				// Log error but continue (don't stop on individual request failures)
				log.Printf("User %d request failed: %v", userID, err)
			}

			// Small delay to prevent overwhelming the system
			// Users run at full speed with minimal delay
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// UseCaseInfo holds use case details
type UseCaseInfo struct {
	ID          string
	Name        string
	Path        string
	Method      string
	Description string
}

// getUseCaseByID returns use case information by ID
func getUseCaseByID(useCaseID string) (UseCaseInfo, error) {
	// Hardcoded use cases matching target-http service
	useCases := map[string]UseCaseInfo{
		"uc1": {
			ID:          "uc1",
			Name:        "Health Check",
			Path:        "/health",
			Method:      "GET",
			Description: "Minimal overhead health check endpoint",
		},
		"uc2": {
			ID:          "uc2",
			Name:        "Status Endpoint",
			Path:        "/api/v1/status",
			Method:      "GET",
			Description: "Status endpoint with lightweight JSON metadata",
		},
		"uc3": {
			ID:          "uc3",
			Name:        "Simple Query",
			Path:        "/api/v1/users/",
			Method:      "GET",
			Description: "Simple query operation with small payload (1-2KB)",
		},
		"uc4": {
			ID:          "uc4",
			Name:        "Create Resource",
			Path:        "/api/v1/users",
			Method:      "POST",
			Description: "Create resource operation with request/response (1-2KB)",
		},
		"uc5": {
			ID:          "uc5",
			Name:        "List Collection",
			Path:        "/api/v1/users",
			Method:      "GET",
			Description: "List/collection query with pagination (50-100KB)",
		},
		"uc6": {
			ID:          "uc6",
			Name:        "Bulk Update",
			Path:        "/api/v1/users/bulk",
			Method:      "PATCH",
			Description: "Bulk update operation with larger payloads (100-200KB)",
		},
		"uc7": {
			ID:          "uc7",
			Name:        "Report Generation",
			Path:        "/api/v1/reports/generate",
			Method:      "POST",
			Description: "CPU-intensive report generation (500KB-1MB response)",
		},
		"uc8": {
			ID:          "uc8",
			Name:        "File Upload",
			Path:        "/api/v1/documents/upload",
			Method:      "POST",
			Description: "Large request payload simulation (10-25MB)",
		},
		"uc9": {
			ID:          "uc9",
			Name:        "Data Export",
			Path:        "/api/v1/transactions/export",
			Method:      "GET",
			Description: "Large response payload and streaming (25-50MB)",
		},
		"uc10": {
			ID:          "uc10",
			Name:        "Network Saturation",
			Path:        "/api/v1/data/process",
			Method:      "POST",
			Description: "Network bandwidth saturation testing (100-250MB)",
		},
		"uc11": {
			ID:          "uc11",
			Name:        "Memory Pressure",
			Path:        "/api/v1/analytics/stream",
			Method:      "POST",
			Description: "Long-running operations for memory leak detection (10-20KB)",
		},
	}

	useCase, exists := useCases[useCaseID]
	if !exists {
		return UseCaseInfo{}, fmt.Errorf("use case %s not found", useCaseID)
	}

	return useCase, nil
}

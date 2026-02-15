package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"time"

	"mallekoppie/ChaosGenerator/internal/contracts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCExecutor handles gRPC requests to target services
type GRPCExecutor struct {
	conn             *grpc.ClientConn
	client           contracts.ChaosTargetClient
	targetAddress    string
	connectionPooled bool
	useCaseID        string
	testExecutionID  string
	sni              string
}

// NewGRPCExecutor creates a new gRPC executor
func NewGRPCExecutor(targetAddress, targetProtocol string, connectionPooled bool, useCaseID, testExecutionID, sni string) (*GRPCExecutor, error) {
	// Configure connection options
	var opts []grpc.DialOption

	// Configure TLS if using grpc protocol (assuming TLS)
	if targetProtocol == "grpc" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // For testing purposes
		}
		if sni != "" {
			tlsConfig.ServerName = sni
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		// Use insecure connection for testing
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Set max message sizes to handle large payloads
	opts = append(opts,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(256*1024*1024), // 256MB
			grpc.MaxCallSendMsgSize(256*1024*1024), // 256MB
		),
	)

	// Connection pooling control
	if !connectionPooled {
		// For non-pooled connections, we'll create a new connection for each execution
		// This is handled in the Execute method
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.WaitForReady(false)))
	}

	// Create connection
	conn, err := grpc.Dial(targetAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Create client
	client := contracts.NewChaosTargetClient(conn)

	return &GRPCExecutor{
		conn:             conn,
		client:           client,
		targetAddress:    targetAddress,
		connectionPooled: connectionPooled,
		useCaseID:        useCaseID,
		testExecutionID:  testExecutionID,
		sni:              sni,
	}, nil
}

// Execute performs a single gRPC request based on the use case and records metrics
func (e *GRPCExecutor) Execute(ctx context.Context) error {
	start := time.Now()

	// If connection pooling is disabled, create a new connection for each request
	if !e.connectionPooled {
		if err := e.reconnect(); err != nil {
			AgentErrorsTotal.WithLabelValues(e.useCaseID, "connection_failed", fmt.Sprintf("%v", e.connectionPooled)).Inc()
			return fmt.Errorf("failed to reconnect: %w", err)
		}
		defer e.conn.Close()
	}

	// Execute the appropriate use case
	var err error
	var method string

	switch e.useCaseID {
	case "uc1":
		method = "HealthCheck"
		err = e.executeHealthCheck(ctx)
	case "uc2":
		method = "GetStatus"
		err = e.executeGetStatus(ctx)
	case "uc3":
		method = "GetUser"
		err = e.executeGetUser(ctx)
	case "uc4":
		method = "CreateUser"
		err = e.executeCreateUser(ctx)
	case "uc5":
		method = "ListUsers"
		err = e.executeListUsers(ctx)
	case "uc6":
		method = "BulkUpdate"
		err = e.executeBulkUpdate(ctx)
	case "uc7":
		method = "GenerateReport"
		err = e.executeGenerateReport(ctx)
	case "uc8":
		method = "UploadDocument"
		err = e.executeUploadDocument(ctx)
	case "uc9":
		method = "ExportData"
		err = e.executeExportData(ctx)
	case "uc10":
		method = "ProcessData"
		err = e.executeProcessData(ctx)
	case "uc11":
		method = "StreamAnalytics"
		err = e.executeStreamAnalytics(ctx)
	default:
		return fmt.Errorf("unknown use case: %s", e.useCaseID)
	}

	// Record metrics
	duration := time.Since(start).Seconds()

	if err != nil {
		AgentErrorsTotal.WithLabelValues(e.useCaseID, "request_failed", fmt.Sprintf("%v", e.connectionPooled)).Inc()
		AgentRequestDuration.WithLabelValues(e.useCaseID, method, "error", fmt.Sprintf("%v", e.connectionPooled)).Observe(duration)
		AgentRequestsTotal.WithLabelValues(e.useCaseID, method, "error", fmt.Sprintf("%v", e.connectionPooled)).Inc()
		return fmt.Errorf("request failed: %w", err)
	}

	AgentRequestDuration.WithLabelValues(e.useCaseID, method, "OK", fmt.Sprintf("%v", e.connectionPooled)).Observe(duration)
	AgentProcessingTime.WithLabelValues(e.useCaseID, fmt.Sprintf("%v", e.connectionPooled)).Observe(duration)
	AgentRequestsTotal.WithLabelValues(e.useCaseID, method, "OK", fmt.Sprintf("%v", e.connectionPooled)).Inc()
	AgentSuccessTotal.WithLabelValues(e.useCaseID, fmt.Sprintf("%v", e.connectionPooled)).Inc()

	return nil
}

// reconnect creates a new connection (for non-pooled mode)
func (e *GRPCExecutor) reconnect() error {
	if e.conn != nil {
		e.conn.Close()
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(256*1024*1024),
			grpc.MaxCallSendMsgSize(256*1024*1024),
		),
	)

	conn, err := grpc.Dial(e.targetAddress, opts...)
	if err != nil {
		return err
	}

	e.conn = conn
	e.client = contracts.NewChaosTargetClient(conn)
	return nil
}

// Use case implementations

func (e *GRPCExecutor) executeHealthCheck(ctx context.Context) error {
	_, err := e.client.HealthCheck(ctx, &contracts.HealthCheckRequest{})
	return err
}

func (e *GRPCExecutor) executeGetStatus(ctx context.Context) error {
	_, err := e.client.GetStatus(ctx, &contracts.GetStatusRequest{})
	return err
}

func (e *GRPCExecutor) executeGetUser(ctx context.Context) error {
	_, err := e.client.GetUser(ctx, &contracts.GetUserRequest{
		UserId: "user-12345",
	})
	return err
}

func (e *GRPCExecutor) executeCreateUser(ctx context.Context) error {
	_, err := e.client.CreateUser(ctx, &contracts.CreateUserRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		FirstName: "Jane",
		LastName:  "Smith",
		Password:  "encrypted_hash",
		Roles:     []string{"user"},
		Preferences: &contracts.UserPreferences{
			Theme:    "light",
			Language: "en",
		},
	})
	return err
}

func (e *GRPCExecutor) executeListUsers(ctx context.Context) error {
	_, err := e.client.ListUsers(ctx, &contracts.ListUsersRequest{
		Page:   1,
		Limit:  50,
		Filter: "active",
	})
	return err
}

func (e *GRPCExecutor) executeBulkUpdate(ctx context.Context) error {
	// Generate 100 update operations
	operations := make([]*contracts.UpdateOperation, 100)
	for i := 0; i < 100; i++ {
		operations[i] = &contracts.UpdateOperation{
			Id: fmt.Sprintf("user-%d", i),
			Fields: map[string]string{
				"lastLogin":         time.Now().Format(time.RFC3339),
				"preferences.theme": "dark",
			},
		}
	}

	_, err := e.client.BulkUpdate(ctx, &contracts.BulkUpdateRequest{
		Updates: operations,
		Options: &contracts.BulkUpdateOptions{
			ValidateAll: true,
			Atomic:      false,
		},
	})
	return err
}

func (e *GRPCExecutor) executeGenerateReport(ctx context.Context) error {
	_, err := e.client.GenerateReport(ctx, &contracts.GenerateReportRequest{
		ReportType: "transactionSummary",
		DateRange: &contracts.DateRange{
			Start: "2026-01-01",
			End:   "2026-02-14",
		},
		Filters: &contracts.ReportFilters{
			Status:    []string{"completed", "pending"},
			MinAmount: 100.0,
		},
		Format:         "json",
		GroupBy:        []string{"date", "category"},
		IncludeDetails: true,
	})
	return err
}

func (e *GRPCExecutor) executeUploadDocument(ctx context.Context) error {
	// Generate 10MB payload
	data := make([]byte, 10*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	_, err := e.client.UploadDocument(ctx, &contracts.UploadDocumentRequest{
		Filename:    "document.pdf",
		ContentType: "application/pdf",
		Size:        int64(len(data)),
		Metadata: &contracts.DocumentMetadata{
			Author:  "John Doe",
			Created: time.Now().Format(time.RFC3339),
			Tags:    []string{"invoice", "q1-2026"},
		},
		Data:     data,
		Checksum: "sha256_hash_here",
	})
	return err
}

func (e *GRPCExecutor) executeExportData(ctx context.Context) error {
	stream, err := e.client.ExportData(ctx, &contracts.ExportDataRequest{
		StartDate: "2026-01-01",
		EndDate:   "2026-02-14",
		Format:    "json",
	})
	if err != nil {
		return err
	}

	// Read all chunks from stream
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		_ = chunk // Process chunk (discard for now)
	}

	return nil
}

func (e *GRPCExecutor) executeProcessData(ctx context.Context) error {
	// Generate 1000 data records (~100MB payload)
	recordCount := 1000
	records := make([]*contracts.DataRecord, recordCount)

	for i := 0; i < recordCount; i++ {
		records[i] = &contracts.DataRecord{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			SensorId:  fmt.Sprintf("sensor-%d", i%100),
			Values: &contracts.SensorValues{
				Temperature:    22.5,
				Humidity:       45.2,
				Pressure:       1013.25,
				WindSpeed:      12.3,
				WindDirection:  180.0,
				Precipitation:  0.0,
				SolarRadiation: 450.5,
				UvIndex:        3.2,
				AirQuality:     85.0,
				NoiseLevel:     45.6,
			},
			Metadata: &contracts.RecordMetadata{
				Location: &contracts.Location{
					Lat: 40.7128,
					Lon: -74.0060,
				},
				Quality:    "high",
				Calibrated: true,
			},
		}
	}

	_, err := e.client.ProcessData(ctx, &contracts.ProcessDataRequest{
		OperationId: fmt.Sprintf("op-%d", time.Now().UnixNano()),
		DatasetType: "timeseries",
		Compression: "none",
		RecordCount: int32(recordCount),
		Records:     records,
	})
	return err
}

func (e *GRPCExecutor) executeStreamAnalytics(ctx context.Context) error {
	_, err := e.client.StreamAnalytics(ctx, &contracts.StreamAnalyticsRequest{
		StreamId:     "stream-12345",
		WindowSize:   60,
		Aggregations: []string{"count", "sum", "average", "min", "max"},
		Buffer: &contracts.BufferConfig{
			Size:     10000,
			Strategy: "sliding",
		},
		Metrics: []string{"request_rate", "error_rate", "latency_p50", "latency_p99"},
	})
	return err
}

// Close closes the gRPC connection
func (e *GRPCExecutor) Close() error {
	if e.conn != nil {
		return e.conn.Close()
	}
	return nil
}

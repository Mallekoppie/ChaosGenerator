package service

import (
	"sync"
	"time"

	"mallekoppie/ChaosGenerator/internal/contracts"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

// RunningTest represents a currently running test execution
type RunningTest struct {
	TestExecutionId        string
	UseCaseId              string
	UseCaseName            string
	TargetId               string
	TargetName             string
	SimulatedUsersPerAgent int32
	AgentIds               []string
	StartTime              time.Time
}

// TestExecutionTracker tracks all running test executions
type TestExecutionTracker struct {
	mu    sync.RWMutex
	tests map[string]*RunningTest
}

var (
	tracker     *TestExecutionTracker
	trackerOnce sync.Once
)

// GetTestExecutionTracker returns the singleton test execution tracker
func GetTestExecutionTracker() *TestExecutionTracker {
	trackerOnce.Do(func() {
		tracker = &TestExecutionTracker{
			tests: make(map[string]*RunningTest),
		}
	})
	return tracker
}

// AddTest adds a new running test
func (t *TestExecutionTracker) AddTest(test *RunningTest) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.tests[test.TestExecutionId] = test
	platform.Log.Info("Test execution added to tracker",
		zap.String("testExecutionId", test.TestExecutionId),
		zap.String("useCaseId", test.UseCaseId),
		zap.Int("agents", len(test.AgentIds)))
}

// RemoveTest removes a test by ID
func (t *TestExecutionTracker) RemoveTest(testExecutionId string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.tests[testExecutionId]; exists {
		delete(t.tests, testExecutionId)
		platform.Log.Info("Test execution removed from tracker",
			zap.String("testExecutionId", testExecutionId))
	}
}

// GetTest gets a test by ID
func (t *TestExecutionTracker) GetTest(testExecutionId string) (*RunningTest, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	test, ok := t.tests[testExecutionId]
	return test, ok
}

// GetAllTests returns all running tests
func (t *TestExecutionTracker) GetAllTests() []*RunningTest {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tests := make([]*RunningTest, 0, len(t.tests))
	for _, test := range t.tests {
		tests = append(tests, test)
	}
	return tests
}

// ToProto converts a RunningTest to protobuf format
func (rt *RunningTest) ToProto() *contracts.RunningTestExecution {
	return &contracts.RunningTestExecution{
		TestExecutionId:        rt.TestExecutionId,
		UseCaseId:              rt.UseCaseId,
		UseCaseName:            rt.UseCaseName,
		TargetId:               rt.TargetId,
		TargetName:             rt.TargetName,
		SimulatedUsersPerAgent: rt.SimulatedUsersPerAgent,
		NumberOfAgents:         int32(len(rt.AgentIds)),
		StartTime:              rt.StartTime.Unix(),
	}
}

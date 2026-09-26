package repositories

import (
	"encoding/json"
	"sort"

	"mallekoppie/ChaosGenerator/internal/master/models"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

const (
	bucketTestRun string = "testrun"
)

// SaveTestRun persists (or replaces) the snapshot of one finished test run.
func SaveTestRun(run models.TestRunHistory) error {
	if err := platform.Database.BoltDb.SaveObject(bucketTestRun, run.ExecutionId, run); err != nil {
		platform.Log.Error("Error saving test run", zap.String("executionId", run.ExecutionId), zap.Error(err))
		return err
	}

	return nil
}

// GetTestRun loads the persisted snapshot of one finished test run.
func GetTestRun(id string) (models.TestRunHistory, error) {
	run := models.TestRunHistory{}
	if err := platform.Database.BoltDb.ReadObject(bucketTestRun, id, &run); err != nil {
		return run, err
	}

	return run, nil
}

// GetAllTestRuns loads every persisted test run, newest finish first.
func GetAllTestRuns() ([]models.TestRunHistory, error) {
	runs := make([]models.TestRunHistory, 0)

	results, err := platform.Database.BoltDb.ReadAllObjects(bucketTestRun)
	if err != nil && err == platform.ErrNoEntryFoundInDB {
		return runs, nil
	}
	if err != nil {
		platform.Log.Error("Error retrieving test runs", zap.Error(err))
		return nil, err
	}

	for _, value := range results {
		run := models.TestRunHistory{}
		if err := json.Unmarshal([]byte(value), &run); err != nil {
			platform.Log.Error("Error unmarshalling test run", zap.Error(err))
			continue
		}

		runs = append(runs, run)
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].EndTime.After(runs[j].EndTime)
	})

	return runs, nil
}

// RemoveTestRun deletes the persisted snapshot of one finished test run.
func RemoveTestRun(id string) error {
	if err := platform.Database.BoltDb.RemoveObject(bucketTestRun, id); err != nil {
		platform.Log.Error("Error removing test run", zap.String("executionId", id), zap.Error(err))
		return err
	}

	return nil
}

// RemoveAllTestRuns deletes every persisted test run.
func RemoveAllTestRuns() error {
	if err := platform.Database.BoltDb.RemoveBucket(bucketTestRun); err != nil {
		platform.Log.Error("Error removing all test runs", zap.Error(err))
		return err
	}

	return nil
}

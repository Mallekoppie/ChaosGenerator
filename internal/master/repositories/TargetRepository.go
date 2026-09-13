package repositories

import (
	"encoding/json"
	"errors"
	"time"

	"mallekoppie/ChaosGenerator/internal/master/models"

	"github.com/Mallekoppie/goslow/platform"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	bucketTarget string = "target"
)

var (
	ErrTargetNotFound      = errors.New("target not found")
	ErrTargetAlreadyExists = errors.New("target already exists")
)

// AddTarget adds a new target to the database
func AddTarget(target models.Target) error {
	// Generate ID if not provided
	if target.ID == "" {
		target.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	target.CreatedAt = now
	target.UpdatedAt = now

	err := platform.Database.BoltDb.SaveObject(bucketTarget, target.ID, target)
	if err != nil {
		platform.Log.Error("Error saving target to DB", zap.Error(err))
		return err
	}

	return nil
}

// GetTarget retrieves a target by ID
func GetTarget(id string) (models.Target, error) {
	var target models.Target
	err := platform.Database.BoltDb.ReadObject(bucketTarget, id, &target)
	if err != nil {
		if err == platform.ErrNoEntryFoundInDB {
			return target, ErrTargetNotFound
		}
		platform.Log.Error("Error reading target from database", zap.String("id", id), zap.Error(err))
		return target, err
	}

	return target, nil
}

// GetAllTargets retrieves all targets from the database
func GetAllTargets() ([]models.Target, error) {
	objects, err := platform.Database.BoltDb.ReadAllObjects(bucketTarget)
	if err != nil {
		platform.Log.Error("Error reading all targets from database", zap.Error(err))
		return nil, err
	}

	targets := make([]models.Target, 0)
	for _, v := range objects {
		var target models.Target
		err = json.Unmarshal([]byte(v), &target)
		if err != nil {
			platform.Log.Error("Error unmarshalling target json data", zap.Error(err))
			continue
		}
		targets = append(targets, target)
	}

	return targets, nil
}

// UpdateTarget updates an existing target
func UpdateTarget(target models.Target) error {
	// Check if target exists
	existingTarget, err := GetTarget(target.ID)
	if err != nil {
		return err
	}

	// Preserve creation time and update timestamp
	target.CreatedAt = existingTarget.CreatedAt
	target.UpdatedAt = time.Now()

	err = platform.Database.BoltDb.SaveObject(bucketTarget, target.ID, target)
	if err != nil {
		platform.Log.Error("Error updating target in DB", zap.Error(err))
		return err
	}

	return nil
}

// DeleteTarget deletes a target by ID
func DeleteTarget(id string) error {
	// Check if target exists
	_, err := GetTarget(id)
	if err != nil {
		return err
	}

	err = platform.Database.BoltDb.RemoveObject(bucketTarget, id)
	if err != nil {
		platform.Log.Error("Error deleting target from DB", zap.String("id", id), zap.Error(err))
		return err
	}

	return nil
}

// DoesTargetExist checks if a target exists by ID
func DoesTargetExist(id string) (bool, error) {
	var target models.Target
	err := platform.Database.BoltDb.ReadObject(bucketTarget, id, &target)
	if err != nil && err == platform.ErrNoEntryFoundInDB {
		return false, nil
	} else if err != nil {
		platform.Log.Error("Error checking if target exists", zap.Error(err))
		return false, err
	}

	return true, nil
}

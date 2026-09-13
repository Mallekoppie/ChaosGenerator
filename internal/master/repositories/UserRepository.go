package repositories

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"mallekoppie/ChaosGenerator/internal/master/models"

	"github.com/Mallekoppie/goslow/platform"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	bucketUser string = "user"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("username already exists")
)

// AddUser persists a new user. The id and creation timestamp are generated when
// they are not supplied by the caller.
func AddUser(user models.User) error {
	if user.Id == "" {
		user.Id = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	existing, err := GetUserByUsername(user.Username)
	if err == nil && existing.Id != "" {
		return ErrUserAlreadyExists
	} else if err != nil && err != ErrUserNotFound {
		return err
	}

	err = platform.Database.BoltDb.SaveObject(bucketUser, user.Id, user)
	if err != nil {
		platform.Log.Error("Error saving user", zap.Error(err))
		return err
	}

	return nil
}

// GetUser retrieves a user by id.
func GetUser(id string) (models.User, error) {
	var user models.User

	err := platform.Database.BoltDb.ReadObject(bucketUser, id, &user)
	if err != nil {
		if err == platform.ErrNoEntryFoundInDB {
			return user, ErrUserNotFound
		}
		platform.Log.Error("Error reading user from database", zap.String("id", id), zap.Error(err))
		return user, err
	}

	if user.Id == "" {
		return user, ErrUserNotFound
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username (case-insensitive).
func GetUserByUsername(username string) (models.User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return models.User{}, err
	}

	for _, user := range users {
		if strings.EqualFold(user.Username, username) {
			return user, nil
		}
	}

	return models.User{}, ErrUserNotFound
}

// GetAllUsers retrieves every user from the database.
func GetAllUsers() ([]models.User, error) {
	users := make([]models.User, 0)

	objects, err := platform.Database.BoltDb.ReadAllObjects(bucketUser)
	if err != nil {
		if err == platform.ErrNoEntryFoundInDB {
			return users, nil
		}
		platform.Log.Error("Error reading all users from database", zap.Error(err))
		return nil, err
	}

	for _, v := range objects {
		var user models.User
		err = json.Unmarshal([]byte(v), &user)
		if err != nil {
			platform.Log.Error("Error unmarshalling user json data", zap.Error(err))
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// DeleteUser deletes a user by id.
func DeleteUser(id string) error {
	_, err := GetUser(id)
	if err != nil {
		return err
	}

	err = platform.Database.BoltDb.RemoveObject(bucketUser, id)
	if err != nil {
		platform.Log.Error("Error deleting user from database", zap.String("id", id), zap.Error(err))
		return err
	}

	return nil
}

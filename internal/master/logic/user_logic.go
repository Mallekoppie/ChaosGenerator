package logic

import (
	"errors"
	"os"
	"strings"

	"mallekoppie/ChaosGenerator/internal/master/models"
	"mallekoppie/ChaosGenerator/internal/master/repositories"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	minUsernameLength         = 3
	minPasswordLength         = 8
	defaultRegisterSecretName = "chaos-dev-secret"
)

var (
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrInvalidRegisterSecret = errors.New("invalid registration secret")
	ErrUsernameTaken         = errors.New("username is already taken")
	ErrWeakCredentials       = errors.New("username must be at least 3 characters and password at least 8 characters")
	ErrLastUserCannotDelete  = errors.New("the last user cannot be deleted")
)

// RegisterSecret returns the secret required to create a user. It is read from
// the CHAOS_REGISTER_SECRET environment variable and falls back to a development
// only value.
func RegisterSecret() string {
	if secret := os.Getenv("CHAOS_REGISTER_SECRET"); secret != "" {
		return secret
	}
	return defaultRegisterSecretName
}

// RegisterUser creates a new user and returns a signed JWT for them.
func RegisterUser(username, password, registerSecret string) (string, error) {
	platform.Log.Info("Registering user", zap.String("username", username))

	username = strings.TrimSpace(username)

	if registerSecret != RegisterSecret() {
		platform.Log.Error("Invalid registration secret supplied", zap.String("username", username))
		return "", ErrInvalidRegisterSecret
	}

	if len(username) < minUsernameLength || len(password) < minPasswordLength {
		return "", ErrWeakCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		platform.Log.Error("Unable to hash password", zap.Error(err))
		return "", err
	}

	user := models.User{
		Username:     username,
		PasswordHash: string(hash),
	}

	err = repositories.AddUser(user)
	if err != nil {
		if err == repositories.ErrUserAlreadyExists {
			return "", ErrUsernameTaken
		}
		platform.Log.Error("Unable to save user", zap.String("username", username), zap.Error(err))
		return "", err
	}

	token, err := newUserToken(username)
	if err != nil {
		return "", err
	}

	platform.Log.Info("User registered successfully", zap.String("username", username))
	return token, nil
}

// LoginUser validates credentials and returns a signed JWT on success.
func LoginUser(username, password string) (string, error) {
	username = strings.TrimSpace(username)

	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		platform.Log.Error("Unable to find user for login", zap.String("username", username), zap.Error(err))
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		platform.Log.Error("Invalid password supplied", zap.String("username", username))
		return "", ErrInvalidCredentials
	}

	token, err := newUserToken(user.Username)
	if err != nil {
		return "", err
	}

	platform.Log.Info("User logged in successfully", zap.String("username", username))
	return token, nil
}

// GetAllUsers returns every registered user.
func GetAllUsers() ([]models.User, error) {
	platform.Log.Info("Getting all users")

	users, err := repositories.GetAllUsers()
	if err != nil {
		platform.Log.Error("Unable to get users", zap.Error(err))
		return nil, err
	}

	return users, nil
}

// DeleteUser removes a user, refusing to remove the final remaining user so the
// system cannot be locked out.
func DeleteUser(id string) error {
	platform.Log.Info("Deleting user", zap.String("id", id))

	users, err := repositories.GetAllUsers()
	if err != nil {
		return err
	}
	if len(users) <= 1 {
		return ErrLastUserCannotDelete
	}

	err = repositories.DeleteUser(id)
	if err != nil {
		platform.Log.Error("Unable to delete user", zap.String("id", id), zap.Error(err))
		return err
	}

	platform.Log.Info("User deleted", zap.String("id", id))
	return nil
}

// newUserToken creates a signed local JWT for the supplied username.
func newUserToken(username string) (string, error) {
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		return "", err
	}

	token, err := platform.LocalJwt.NewLocalJwtToken(map[string]interface{}{
		"sub":      user.Id,
		"username": user.Username,
	})
	if err != nil {
		platform.Log.Error("Unable to create JWT token", zap.String("username", username), zap.Error(err))
		return "", err
	}

	return token, nil
}

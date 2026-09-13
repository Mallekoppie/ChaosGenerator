package logic

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mallekoppie/ChaosGenerator/internal/master/repositories"

	"github.com/Mallekoppie/goslow/platform"
)

const testRegisterSecret = "unit-test-register-secret"

var testDBFile string

func TestMain(m *testing.M) {
	os.Setenv("CHAOS_REGISTER_SECRET", testRegisterSecret)

	testDBFile = filepath.Join(os.TempDir(), fmt.Sprintf("chaos_logic_%d.db", time.Now().UnixNano()))

	config := platform.Config{}
	config.Component.ComponentName = "chaos-logic-tests"
	config.Log.Level = "error"
	config.Auth.Server.LocalJwt.Enabled = true
	config.Auth.Server.LocalJwt.JwtSigningKey = "unit-test-signing-key"
	config.Auth.Server.LocalJwt.JwtSigningMethod = "HS256"
	config.Auth.Server.LocalJwt.JwtExpiration = 60
	config.Database.BoltDB.Enabled = true
	config.Database.BoltDB.FileName = testDBFile

	platform.SetPlatformConfiguration(config)

	if err := platform.SetupBoltDB(); err != nil {
		fmt.Println("Unable to set up BoltDB for tests:", err)
		os.Exit(1)
	}

	code := m.Run()

	_ = platform.Database.BoltDb.Close()
	os.Remove(testDBFile)
	os.Remove(testDBFile + ".lock")

	os.Exit(code)
}

func TestRegisterUserAndLogin(t *testing.T) {
	token, err := RegisterUser("logicuser", "password123", testRegisterSecret)
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}
	if token == "" {
		t.Fatal("Expected a token to be returned from registration")
	}

	claims, err := platform.LocalJwt.ValidateLocalJwtToken(token)
	if err != nil {
		t.Fatalf("Token returned from registration should be valid: %v", err)
	}
	if claims["username"] != "logicuser" {
		t.Fatalf("Unexpected token claims: %v", claims)
	}

	loginToken, err := LoginUser("logicuser", "password123")
	if err != nil {
		t.Fatalf("LoginUser failed: %v", err)
	}
	if loginToken == "" {
		t.Fatal("Expected a token to be returned from login")
	}
}

func TestRegisterUserInvalidSecret(t *testing.T) {
	_, err := RegisterUser("logicbadsecret", "password123", "the-wrong-secret")
	if err != ErrInvalidRegisterSecret {
		t.Fatalf("Expected ErrInvalidRegisterSecret, got %v", err)
	}
}

func TestRegisterUserWeakCredentials(t *testing.T) {
	_, err := RegisterUser("ab", "short", testRegisterSecret)
	if err != ErrWeakCredentials {
		t.Fatalf("Expected ErrWeakCredentials, got %v", err)
	}
}

func TestRegisterUserDuplicateUsername(t *testing.T) {
	if _, err := RegisterUser("logicdupe", "password123", testRegisterSecret); err != nil {
		t.Fatalf("Failed to register first user: %v", err)
	}

	_, err := RegisterUser("logiCDupe", "password123", testRegisterSecret)
	if err != ErrUsernameTaken {
		t.Fatalf("Expected ErrUsernameTaken, got %v", err)
	}
}

func TestLoginUserInvalidPassword(t *testing.T) {
	if _, err := RegisterUser("logicwrongpass", "password123", testRegisterSecret); err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	_, err := LoginUser("logicwrongpass", "not-the-password")
	if err != ErrInvalidCredentials {
		t.Fatalf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUserUnknown(t *testing.T) {
	_, err := LoginUser("lognouser", "password123")
	if err != ErrInvalidCredentials {
		t.Fatalf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	if _, err := RegisterUser("logicdelete", "password123", testRegisterSecret); err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	user, err := repositories.GetUserByUsername("logicdelete")
	if err != nil {
		t.Fatalf("Failed to look up user: %v", err)
	}

	users, err := GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	// Only delete while more than one user remains; deleting the final user is
	// explicitly disallowed.
	if len(users) > 1 {
		if err := DeleteUser(user.Id); err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}
		if _, err := repositories.GetUserByUsername("logicdelete"); err != repositories.ErrUserNotFound {
			t.Fatalf("Expected user to be deleted, got %v", err)
		}
	}

	remaining, err := GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}
	if len(remaining) == 1 {
		if err := DeleteUser(remaining[0].Id); err != ErrLastUserCannotDelete {
			t.Fatalf("Expected ErrLastUserCannotDelete, got %v", err)
		}
	}
}

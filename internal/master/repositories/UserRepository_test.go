package repositories

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mallekoppie/ChaosGenerator/internal/master/models"

	"github.com/Mallekoppie/goslow/platform"
)

var testDBFile string

// TestMain configures the platform with a temporary BoltDB so the repository
// tests (including the pre-existing agent tests) have a working database.
func TestMain(m *testing.M) {
	testDBFile = filepath.Join(os.TempDir(), fmt.Sprintf("chaos_repositories_%d.db", time.Now().UnixNano()))

	config := platform.Config{}
	config.Component.ComponentName = "chaos-repositories-tests"
	config.Log.Level = "error"
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

func TestAddAndGetUser(t *testing.T) {
	user := models.User{
		Id:           "user-add-get",
		Username:     "AddGetUser",
		PasswordHash: "not-a-real-hash",
	}

	if err := AddUser(user); err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	found, err := GetUser("user-add-get")
	if err != nil {
		t.Fatalf("Failed to get user by id: %v", err)
	}
	if found.Username != "AddGetUser" {
		t.Fatalf("Unexpected username: %s", found.Username)
	}
	if found.PasswordHash != "not-a-real-hash" {
		t.Fatalf("Password hash was not persisted")
	}
	if found.CreatedAt.IsZero() {
		t.Fatalf("Expected CreatedAt to be populated")
	}

	byName, err := GetUserByUsername("addgetuser")
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}
	if byName.Id != "user-add-get" {
		t.Fatalf("Unexpected id returned: %s", byName.Id)
	}
}

func TestAddDuplicateUser(t *testing.T) {
	if err := AddUser(models.User{Username: "DuplicateUser", PasswordHash: "hash"}); err != nil {
		t.Fatalf("Failed to add first user: %v", err)
	}

	err := AddUser(models.User{Username: "duplicateuser", PasswordHash: "hash"})
	if err != ErrUserAlreadyExists {
		t.Fatalf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestGetMissingUser(t *testing.T) {
	if _, err := GetUser("user-that-does-not-exist"); err != ErrUserNotFound {
		t.Fatalf("Expected ErrUserNotFound for id lookup, got %v", err)
	}
	if _, err := GetUserByUsername("missing-from-database"); err != ErrUserNotFound {
		t.Fatalf("Expected ErrUserNotFound for username lookup, got %v", err)
	}
}

func TestGetAllAndDeleteUser(t *testing.T) {
	user := models.User{Id: "user-delete", Username: "DeleteThisUser", PasswordHash: "hash"}
	if err := AddUser(user); err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	users, err := GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	found := false
	for _, u := range users {
		if u.Id == "user-delete" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Expected to find the newly created user")
	}

	if err := DeleteUser("user-delete"); err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if _, err := GetUser("user-delete"); err != ErrUserNotFound {
		t.Fatalf("Expected the user to be gone, got %v", err)
	}
}

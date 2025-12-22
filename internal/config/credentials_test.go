package config

import (
	"os"
	"testing"
)

func TestSaveAndGetCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	t.Run("save and retrieve credentials", func(t *testing.T) {
		err := SaveCredentials("server1", "admin", "secret123")
		if err != nil {
			t.Fatalf("SaveCredentials() error = %v", err)
		}

		user, pass, err := GetCredentials("server1")
		if err != nil {
			t.Fatalf("GetCredentials() error = %v", err)
		}
		if user != "admin" {
			t.Errorf("username = %q, want %q", user, "admin")
		}
		if pass != "secret123" {
			t.Errorf("password = %q, want %q", pass, "secret123")
		}
	})

	t.Run("update existing credentials", func(t *testing.T) {
		err := SaveCredentials("server1", "newuser", "newpass")
		if err != nil {
			t.Fatalf("SaveCredentials() error = %v", err)
		}

		user, pass, err := GetCredentials("server1")
		if err != nil {
			t.Fatalf("GetCredentials() error = %v", err)
		}
		if user != "newuser" {
			t.Errorf("username = %q, want %q", user, "newuser")
		}
		if pass != "newpass" {
			t.Errorf("password = %q, want %q", pass, "newpass")
		}
	})

	t.Run("get non-existing credentials", func(t *testing.T) {
		_, _, err := GetCredentials("nonexistent")
		if err == nil {
			t.Error("GetCredentials() should return error for non-existing server")
		}
	})
}

func TestDeleteCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// First save some credentials
	SaveCredentials("todelete", "user", "pass")

	// Delete them
	err := DeleteCredentials("todelete")
	if err != nil {
		t.Fatalf("DeleteCredentials() error = %v", err)
	}

	// Verify they're gone
	_, _, err = GetCredentials("todelete")
	if err == nil {
		t.Error("credentials should be deleted")
	}
}

func TestCredentialsPath(t *testing.T) {
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	path := credentialsPath()
	expected := tmpDir + "/xstream-tui/credentials.json"
	if path != expected {
		t.Errorf("credentialsPath() = %q, want %q", path, expected)
	}
}

func TestLoadCredentials_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// No credentials file exists
	creds, err := loadCredentials()
	if err != nil {
		t.Fatalf("loadCredentials() error = %v", err)
	}
	if creds == nil {
		t.Error("loadCredentials() should return empty map, not nil")
	}
	if len(creds) != 0 {
		t.Errorf("loadCredentials() returned %d items, want 0", len(creds))
	}
}

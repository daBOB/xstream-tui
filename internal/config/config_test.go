package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		oldVal := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", oldVal)

		os.Setenv("XDG_CONFIG_HOME", "/custom/config")
		got := Dir()
		want := "/custom/config/xstream-tui"
		if got != want {
			t.Errorf("Dir() = %q, want %q", got, want)
		}
	})

	// Test without XDG_CONFIG_HOME (falls back to ~/.config)
	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		oldVal := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", oldVal)

		os.Unsetenv("XDG_CONFIG_HOME")
		got := Dir()
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".config", "xstream-tui")
		if got != want {
			t.Errorf("Dir() = %q, want %q", got, want)
		}
	})
}

func TestLoad(t *testing.T) {
	// Create temp directory for test config
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	t.Run("missing config returns defaults", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Player != "mpv" {
			t.Errorf("default Player = %q, want %q", cfg.Player, "mpv")
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		configDir := filepath.Join(tmpDir, "xstream-tui")
		os.MkdirAll(configDir, 0700)

		content := `
default_server = "test"
player = "vlc"

[[servers]]
name = "test"
host = "example.com"
port = 8080
username = "user"
`
		os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(content), 0600)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.DefaultServer != "test" {
			t.Errorf("DefaultServer = %q, want %q", cfg.DefaultServer, "test")
		}
		if cfg.Player != "vlc" {
			t.Errorf("Player = %q, want %q", cfg.Player, "vlc")
		}
		if len(cfg.Servers) != 1 {
			t.Fatalf("Servers count = %d, want 1", len(cfg.Servers))
		}
		if cfg.Servers[0].Host != "example.com" {
			t.Errorf("Server host = %q, want %q", cfg.Servers[0].Host, "example.com")
		}
	})
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := &Config{
		DefaultServer: "myserver",
		Player:        "mpv",
		Servers: []Server{
			{Name: "myserver", Host: "test.com", Port: 80, Username: "admin"},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists with correct permissions
	path := filepath.Join(tmpDir, "xstream-tui", "config.toml")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("file permissions = %o, want 0600", info.Mode().Perm())
	}

	// Reload and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}
	if loaded.DefaultServer != "myserver" {
		t.Errorf("loaded DefaultServer = %q, want %q", loaded.DefaultServer, "myserver")
	}
}

func TestConfig_AddServer(t *testing.T) {
	cfg := &Config{}

	// Add new server
	cfg.AddServer(Server{Name: "server1", Host: "host1.com"})
	if len(cfg.Servers) != 1 {
		t.Errorf("after first add, len = %d, want 1", len(cfg.Servers))
	}

	// Add another server
	cfg.AddServer(Server{Name: "server2", Host: "host2.com"})
	if len(cfg.Servers) != 2 {
		t.Errorf("after second add, len = %d, want 2", len(cfg.Servers))
	}

	// Update existing server
	cfg.AddServer(Server{Name: "server1", Host: "updated.com"})
	if len(cfg.Servers) != 2 {
		t.Errorf("after update, len = %d, want 2", len(cfg.Servers))
	}
	if cfg.Servers[0].Host != "updated.com" {
		t.Errorf("updated host = %q, want %q", cfg.Servers[0].Host, "updated.com")
	}
}

func TestConfig_GetServer(t *testing.T) {
	cfg := &Config{
		Servers: []Server{
			{Name: "server1", Host: "host1.com"},
			{Name: "server2", Host: "host2.com"},
		},
	}

	t.Run("existing server", func(t *testing.T) {
		s := cfg.GetServer("server1")
		if s == nil {
			t.Fatal("GetServer returned nil for existing server")
		}
		if s.Host != "host1.com" {
			t.Errorf("Host = %q, want %q", s.Host, "host1.com")
		}
	})

	t.Run("non-existing server", func(t *testing.T) {
		s := cfg.GetServer("nonexistent")
		if s != nil {
			t.Errorf("GetServer returned %v for non-existing server, want nil", s)
		}
	})
}

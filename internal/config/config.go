// Package config handles application configuration and credential storage.
package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Server represents a saved IPTV server configuration.
type Server struct {
	Name     string `toml:"name"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	// Password stored separately in credentials file
}

// Config holds application settings.
type Config struct {
	DefaultServer string   `toml:"default_server"`
	Servers       []Server `toml:"servers"`
	Player        string   `toml:"player"` // "mpv" or "vlc"
}

// Dir returns the configuration directory path.
// Respects XDG_CONFIG_HOME on Linux/macOS.
func Dir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "xstream-tui")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "xstream-tui")
}

// Load reads config from file or returns defaults.
func Load() (*Config, error) {
	path := filepath.Join(Dir(), "config.toml")

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		if os.IsNotExist(err) {
			return &Config{Player: "mpv"}, nil // Default config
		}
		return nil, err
	}
	return &cfg, nil
}

// Save writes config to file with secure permissions.
func (c *Config) Save() error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path := filepath.Join(dir, "config.toml")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(c)
}

// AddServer adds or updates a server in the config.
func (c *Config) AddServer(server Server) {
	for i, s := range c.Servers {
		if s.Name == server.Name {
			c.Servers[i] = server
			return
		}
	}
	c.Servers = append(c.Servers, server)
}

// GetServer returns a server by name.
func (c *Config) GetServer(name string) *Server {
	for _, s := range c.Servers {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

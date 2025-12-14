package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credentials stored in JSON file with 0600 permissions.
// Note: In production, consider encryption or OS keyring.

type credentials map[string]serverCreds

type serverCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func credentialsPath() string {
	return filepath.Join(Dir(), "credentials.json")
}

// SaveCredentials stores credentials for a server.
func SaveCredentials(serverName, username, password string) error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	creds, err := loadCredentials()
	if err != nil {
		// Only fail on non-trivial errors; missing file is OK
		return fmt.Errorf("load existing credentials: %w", err)
	}
	if creds == nil {
		creds = make(credentials)
	}

	creds[serverName] = serverCreds{
		Username: username,
		Password: password,
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	return os.WriteFile(credentialsPath(), data, 0600)
}

// GetCredentials retrieves credentials for a server.
func GetCredentials(serverName string) (username, password string, err error) {
	creds, err := loadCredentials()
	if err != nil {
		return "", "", err
	}

	c, ok := creds[serverName]
	if !ok {
		return "", "", fmt.Errorf("credentials not found for %s", serverName)
	}

	return c.Username, c.Password, nil
}

// DeleteCredentials removes credentials for a server.
func DeleteCredentials(serverName string) error {
	creds, err := loadCredentials()
	if err != nil {
		return err
	}

	delete(creds, serverName)

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	return os.WriteFile(credentialsPath(), data, 0600)
}

func loadCredentials() (credentials, error) {
	data, err := os.ReadFile(credentialsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return make(credentials), nil
		}
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	return creds, nil
}

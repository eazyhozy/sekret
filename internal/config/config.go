package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const configDir = "sekret"
const configFile = "config.json"
const currentVersion = 1

// KeyEntry represents metadata for a single registered key.
type KeyEntry struct {
	Name    string    `json:"name"`
	EnvVar  string    `json:"env_var"`
	AddedAt time.Time `json:"added_at"`
}

// Config represents the sekret config file structure.
type Config struct {
	Version int        `json:"version"`
	Keys    []KeyEntry `json:"keys"`
}

// configPath returns the path override if set, or the default XDG path.
var configPathOverride string

// SetPath overrides the config directory for testing.
func SetPath(path string) {
	configPathOverride = path
}

func getConfigPath() (string, error) {
	if configPathOverride != "" {
		return filepath.Join(configPathOverride, configFile), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine config directory: %w", err)
	}
	return filepath.Join(dir, configDir, configFile), nil
}

// Load reads the config file. Returns an empty config if the file does not exist.
func Load() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Version: currentVersion, Keys: []KeyEntry{}}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk, creating directories as needed.
func Save(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// AddKey adds a new key entry. Returns an error if the name or env var already exists.
func (c *Config) AddKey(name, envVar string) error {
	for _, k := range c.Keys {
		if k.Name == name {
			return fmt.Errorf("key %q already exists", name)
		}
		if k.EnvVar == envVar {
			return fmt.Errorf("environment variable %q is already used by key %q", envVar, k.Name)
		}
	}
	c.Keys = append(c.Keys, KeyEntry{
		Name:    name,
		EnvVar:  envVar,
		AddedAt: time.Now(),
	})
	return nil
}

// RemoveKey removes a key entry by name. Returns an error if not found.
func (c *Config) RemoveKey(name string) error {
	for i, k := range c.Keys {
		if k.Name == name {
			c.Keys = append(c.Keys[:i], c.Keys[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("key %q not found", name)
}

// FindKey returns the key entry for the given name, or nil if not found.
func (c *Config) FindKey(name string) *KeyEntry {
	for i := range c.Keys {
		if c.Keys[i].Name == name {
			return &c.Keys[i]
		}
	}
	return nil
}

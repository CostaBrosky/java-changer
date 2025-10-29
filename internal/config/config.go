package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the application configuration
type Config struct {
	CustomPaths []string `json:"custom_paths"` // Specific Java installation paths
	SearchPaths []string `json:"search_paths"` // Base directories to scan for Java installations
	configPath  string
}

// Load loads the configuration from the user's home directory
func Load() (*Config, error) {
	configPath := getConfigPath()

	cfg := &Config{
		CustomPaths: make([]string, 0),
		SearchPaths: make([]string, 0),
		configPath:  configPath,
	}

	// If config file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	cfg.configPath = configPath
	return cfg, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	// Ensure config directory exists
	configDir := filepath.Dir(c.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(c.configPath, data, 0644)
}

// AddCustomPath adds a custom Java installation path
func (c *Config) AddCustomPath(path string) {
	// Normalize path
	path = filepath.Clean(path)

	// Check if already exists
	for _, p := range c.CustomPaths {
		if strings.EqualFold(p, path) {
			return
		}
	}

	c.CustomPaths = append(c.CustomPaths, path)
}

// RemoveCustomPath removes a custom Java installation path
func (c *Config) RemoveCustomPath(path string) {
	path = filepath.Clean(path)

	for i, p := range c.CustomPaths {
		if strings.EqualFold(p, path) {
			c.CustomPaths = append(c.CustomPaths[:i], c.CustomPaths[i+1:]...)
			return
		}
	}
}

// HasCustomPath checks if a path exists in custom paths
func (c *Config) HasCustomPath(path string) bool {
	path = filepath.Clean(path)

	for _, p := range c.CustomPaths {
		if strings.EqualFold(p, path) {
			return true
		}
	}
	return false
}

// AddSearchPath adds a search path for auto-detection
func (c *Config) AddSearchPath(path string) {
	// Normalize path
	path = filepath.Clean(path)

	// Check if already exists
	for _, p := range c.SearchPaths {
		if strings.EqualFold(p, path) {
			return
		}
	}

	c.SearchPaths = append(c.SearchPaths, path)
}

// RemoveSearchPath removes a search path
func (c *Config) RemoveSearchPath(path string) {
	path = filepath.Clean(path)

	for i, p := range c.SearchPaths {
		if strings.EqualFold(p, path) {
			c.SearchPaths = append(c.SearchPaths[:i], c.SearchPaths[i+1:]...)
			return
		}
	}
}

// HasSearchPath checks if a path exists in search paths
func (c *Config) HasSearchPath(path string) bool {
	path = filepath.Clean(path)

	for _, p := range c.SearchPaths {
		if strings.EqualFold(p, path) {
			return true
		}
	}
	return false
}

// getConfigPath returns the path to the configuration file
// Following XDG Base Directory specification
func getConfigPath() string {
	// Try XDG_CONFIG_HOME first (standard on Unix systems)
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome != "" {
		return filepath.Join(configHome, "jv", "jv.json")
	}

	// Fallback to $HOME/.config/jv/jv.json (XDG default)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return filepath.Join(homeDir, ".config", "jv", "jv.json")
}

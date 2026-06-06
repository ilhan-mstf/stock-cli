package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	APIKey    string   `json:"api_key"`
	Watchlist []string `json:"watchlist"`
}

// DefaultPath returns the standard configuration file path for the CLI.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "stock-cli", "config.json"), nil
}

// Load loads the configuration from the specified path.
// If path is empty, it uses the default configuration path.
// If the file does not exist, a default empty configuration is returned.
func Load(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			Watchlist: []string{},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Watchlist == nil {
		cfg.Watchlist = []string{}
	}

	return &cfg, nil
}

// Save writes the configuration to the specified path.
// If path is empty, it uses the default configuration path.
// It creates any missing parent directories and writes with 0600 permissions.
func Save(path string, cfg *Config) error {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return err
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// AddSymbol adds a symbol to the watchlist if not already present.
// It normalizes the symbol to uppercase and returns true if added, false if it was already present.
func (cfg *Config) AddSymbol(sym string) bool {
	sym = strings.ToUpper(strings.TrimSpace(sym))
	if sym == "" {
		return false
	}
	for _, s := range cfg.Watchlist {
		if s == sym {
			return false
		}
	}
	cfg.Watchlist = append(cfg.Watchlist, sym)
	return true
}

// RemoveSymbol removes a symbol from the watchlist.
// It normalizes the symbol to uppercase and returns true if removed, false if it was not present.
func (cfg *Config) RemoveSymbol(sym string) bool {
	sym = strings.ToUpper(strings.TrimSpace(sym))
	for i, s := range cfg.Watchlist {
		if s == sym {
			cfg.Watchlist = append(cfg.Watchlist[:i], cfg.Watchlist[i+1:]...)
			return true
		}
	}
	return false
}

// MaskAPIKey returns a masked representation of the API key for security.
func MaskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	length := len(key)
	if length <= 2 {
		return strings.Repeat("*", length)
	}
	return string(key[0]) + strings.Repeat("*", length-2) + string(key[length-1])
}

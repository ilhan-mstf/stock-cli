package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "*"},
		{"ab", "**"},
		{"abc", "a*c"},
		{"abcd", "a**d"},
		{"mysecretkey", "m*********y"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			actual := MaskAPIKey(tc.input)
			if actual != tc.expected {
				t.Errorf("MaskAPIKey(%q) = %q; want %q", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestWatchlistOperations(t *testing.T) {
	cfg := &Config{
		Watchlist: []string{},
	}

	// Test adding symbol
	if !cfg.AddSymbol("aapl") {
		t.Errorf("Expected AddSymbol(aapl) to return true")
	}
	if !reflect.DeepEqual(cfg.Watchlist, []string{"AAPL"}) {
		t.Errorf("Watchlist = %v; want [AAPL]", cfg.Watchlist)
	}

	// Test adding duplicate
	if cfg.AddSymbol("AAPL") {
		t.Errorf("Expected AddSymbol(AAPL) duplicate to return false")
	}
	if !reflect.DeepEqual(cfg.Watchlist, []string{"AAPL"}) {
		t.Errorf("Watchlist = %v; want [AAPL]", cfg.Watchlist)
	}

	// Test adding empty/whitespace
	if cfg.AddSymbol("  ") {
		t.Errorf("Expected AddSymbol with only spaces to return false")
	}

	// Test adding another symbol
	if !cfg.AddSymbol(" msft ") {
		t.Errorf("Expected AddSymbol( msft ) to return true")
	}
	if !reflect.DeepEqual(cfg.Watchlist, []string{"AAPL", "MSFT"}) {
		t.Errorf("Watchlist = %v; want [AAPL, MSFT]", cfg.Watchlist)
	}

	// Test removing non-existent
	if cfg.RemoveSymbol("GOOG") {
		t.Errorf("Expected RemoveSymbol(GOOG) to return false")
	}

	// Test removing existent
	if !cfg.RemoveSymbol("aapl") {
		t.Errorf("Expected RemoveSymbol(aapl) to return true")
	}
	if !reflect.DeepEqual(cfg.Watchlist, []string{"MSFT"}) {
		t.Errorf("Watchlist = %v; want [MSFT]", cfg.Watchlist)
	}
}

func TestLoadSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stock-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "config.json")

	// Test loading non-existent file returns default empty config
	cfg, err := Load(tempFile)
	if err != nil {
		t.Fatalf("Load non-existent config failed: %v", err)
	}
	if cfg.APIKey != "" || len(cfg.Watchlist) != 0 {
		t.Errorf("Load on non-existent config = %+v; want empty", cfg)
	}

	// Test saving config
	cfg.APIKey = "dummy_key"
	cfg.AddSymbol("AAPL")
	err = Save(tempFile, cfg)
	if err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(tempFile)
	if err != nil {
		t.Fatalf("Stat config failed: %v", err)
	}
	// On UNIX-like systems, verify 0600 (owner read/write only).
	// We mask with 0777 to check the lowest 9 bits.
	// Since os.WriteFile is used, permissions should be 0600.
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Saved config permissions = %v; want 0600", perm)
	}

	// Test loading existing config
	loadedCfg, err := Load(tempFile)
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}
	if loadedCfg.APIKey != "dummy_key" {
		t.Errorf("Loaded APIKey = %q; want 'dummy_key'", loadedCfg.APIKey)
	}
	if !reflect.DeepEqual(loadedCfg.Watchlist, []string{"AAPL"}) {
		t.Errorf("Loaded Watchlist = %v; want [AAPL]", loadedCfg.Watchlist)
	}
}

func TestLoadCorruptedConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stock-cli-corrupt-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "corrupt_config.json")

	// Write corrupted JSON data
	err = os.WriteFile(tempFile, []byte("{invalid_json: true"), 0600)
	if err != nil {
		t.Fatalf("Failed to write corrupted config file: %v", err)
	}

	_, err = Load(tempFile)
	if err == nil {
		t.Error("Expected error when loading corrupted config file, got nil")
	}
}

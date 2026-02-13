package config

import (
	"os"
	"path/filepath"
	"testing"

	"network-log-formatter/internal/model"

	"pgregory.net/rapid"
)

// Feature: network-log-formatter, Property 12: 设置保存/加载往返
// For any valid Settings object, saving then loading should return an equivalent Settings.
// **Validates: Requirements 6.4, 6.5**
func TestProperty12_SettingsSaveLoadRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		original := model.Settings{
			LLM: model.LLMConfig{
				BaseURL:   rapid.String().Draw(t, "baseURL"),
				APIKey:    rapid.String().Draw(t, "apiKey"),
				ModelName: rapid.String().Draw(t, "modelName"),
			},
			UvPath:           rapid.String().Draw(t, "uvPath"),
			DefaultInputDir:  rapid.String().Draw(t, "defaultInputDir"),
			DefaultOutputDir: rapid.String().Draw(t, "defaultOutputDir"),
		}

		tmpDir, err := os.MkdirTemp("", "settings-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		configPath := filepath.Join(tmpDir, "settings.json")
		sm := NewSettingsManager(configPath)

		if err := sm.Save(original); err != nil {
			t.Fatalf("failed to save settings: %v", err)
		}

		loaded, err := sm.Load()
		if err != nil {
			t.Fatalf("failed to load settings: %v", err)
		}

		if original.LLM.BaseURL != loaded.LLM.BaseURL {
			t.Fatalf("LLM.BaseURL mismatch: got %q, want %q", loaded.LLM.BaseURL, original.LLM.BaseURL)
		}
		if original.LLM.APIKey != loaded.LLM.APIKey {
			t.Fatalf("LLM.APIKey mismatch: got %q, want %q", loaded.LLM.APIKey, original.LLM.APIKey)
		}
		if original.LLM.ModelName != loaded.LLM.ModelName {
			t.Fatalf("LLM.ModelName mismatch: got %q, want %q", loaded.LLM.ModelName, original.LLM.ModelName)
		}
		if original.UvPath != loaded.UvPath {
			t.Fatalf("UvPath mismatch: got %q, want %q", loaded.UvPath, original.UvPath)
		}
		if original.DefaultInputDir != loaded.DefaultInputDir {
			t.Fatalf("DefaultInputDir mismatch: got %q, want %q", loaded.DefaultInputDir, original.DefaultInputDir)
		}
		if original.DefaultOutputDir != loaded.DefaultOutputDir {
			t.Fatalf("DefaultOutputDir mismatch: got %q, want %q", loaded.DefaultOutputDir, original.DefaultOutputDir)
		}
	})
}

// TestSettingsManager_LoadNonExistent verifies that loading from a non-existent file returns defaults.
func TestSettingsManager_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", "settings.json")
	sm := NewSettingsManager(configPath)

	settings, err := sm.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	defaults := defaultSettings()
	assertSettingsEqual(t, settings, &defaults)
}

// TestSettingsManager_LoadCorruptedFile verifies that a corrupted config file returns an error.
func TestSettingsManager_LoadCorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "settings.json")

	err := os.WriteFile(configPath, []byte("{invalid json!!!"), 0o644)
	if err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}

	sm := NewSettingsManager(configPath)
	_, err = sm.Load()
	if err == nil {
		t.Fatal("expected error for corrupted settings file, got nil")
	}
}

func assertSettingsEqual(t *testing.T, got, want *model.Settings) {
	t.Helper()
	if got.LLM != want.LLM {
		t.Fatalf("LLM mismatch: got %+v, want %+v", got.LLM, want.LLM)
	}
	if got.UvPath != want.UvPath {
		t.Fatalf("UvPath mismatch: got %q, want %q", got.UvPath, want.UvPath)
	}
	if got.DefaultInputDir != want.DefaultInputDir {
		t.Fatalf("DefaultInputDir mismatch")
	}
	if got.DefaultOutputDir != want.DefaultOutputDir {
		t.Fatalf("DefaultOutputDir mismatch")
	}
}

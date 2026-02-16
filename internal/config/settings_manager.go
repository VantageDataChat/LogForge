package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"network-log-formatter/internal/model"
)

// SettingsManager handles loading and saving application settings
// including LLM configuration, Python environment paths, and default directories.
type SettingsManager struct {
	configPath string
}

// NewSettingsManager creates a new SettingsManager that persists settings to the given file path.
func NewSettingsManager(configPath string) *SettingsManager {
	return &SettingsManager{configPath: configPath}
}

// defaultSettings returns the default application settings.
func defaultSettings() model.Settings {
	showWizard := true
	return model.Settings{
		LLM: model.LLMConfig{
			BaseURL:   "",
			APIKey:    "",
			ModelName: "",
		},
		UvPath:           "uv",
		DefaultInputDir:  "",
		DefaultOutputDir: "",
		SampleLines:      20,
		ShowWizard:       &showWizard,
	}
}

// Load reads settings from the JSON config file.
// If the file does not exist, it returns default settings.
// Other errors (permissions, corrupted JSON) are returned to the caller.
func (sm *SettingsManager) Load() (*model.Settings, error) {
	data, err := os.ReadFile(sm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			defaults := defaultSettings()
			return &defaults, nil
		}
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings model.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("settings file is corrupted: %w", err)
	}

	return &settings, nil
}

// Save writes the given settings to the JSON config file with indented formatting.
// It creates parent directories if they don't exist.
func (sm *SettingsManager) Save(settings model.Settings) error {
	dir := filepath.Dir(sm.configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.configPath, data, 0o644)
}

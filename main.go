package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	configDir, err := getConfigDir()
	if err != nil {
		fmt.Printf("Error determining config directory: %v\n", err)
		os.Exit(1)
	}

	app := NewApp(configDir)

	err = wails.Run(&options.App{
		Title:  "LogForge - 智能日志格式化",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// getConfigDir returns the application configuration directory.
// Uses os.UserConfigDir() and appends the app name subdirectory.
// Falls back to a local ".config" directory if UserConfigDir fails.
func getConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to current directory
		return filepath.Join(".", ".network-log-formatter"), nil
	}
	configDir := filepath.Join(userConfigDir, "network-log-formatter")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return configDir, nil
}

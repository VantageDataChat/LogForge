package pyenv

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// EnvStatus holds the current state of the Python environment.
type EnvStatus struct {
	UvAvailable bool   `json:"uv_available"`
	EnvExists   bool   `json:"env_exists"`
	EnvPath     string `json:"env_path"`
}

// PythonEnvManager manages isolated Python environments using uv.
// Handles environment creation, dependency installation, and script execution.
type PythonEnvManager struct {
	uvPath  string
	envPath string
}

// NewPythonEnvManager creates a new PythonEnvManager with the given uv binary path
// and virtual environment directory path.
func NewPythonEnvManager(uvPath string, envPath string) *PythonEnvManager {
	return &PythonEnvManager{
		uvPath:  uvPath,
		envPath: envPath,
	}
}

// EnsureEnv checks that uv is available, creates the virtual environment if it
// does not exist, and installs required dependencies (openpyxl).
func (pem *PythonEnvManager) EnsureEnv(ctx context.Context) error {
	// 1. Check uv availability
	if err := pem.checkUv(ctx); err != nil {
		return fmt.Errorf("uv is not available: %w. Please install uv (https://docs.astral.sh/uv/)", err)
	}

	// 2. Create virtual environment if it doesn't exist
	if _, err := os.Stat(pem.envPath); os.IsNotExist(err) {
		cmd := exec.CommandContext(ctx, pem.uvPath, "venv", pem.envPath)
		hideWindow(cmd)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create virtual environment: %w\n%s", err, string(output))
		}
	}

	// 3. Install dependencies (only if not already installed)
	pythonBin := pem.pythonPath()
	checkCmd := exec.CommandContext(ctx, pythonBin, "-c", "import openpyxl")
	hideWindow(checkCmd)
	if checkCmd.Run() != nil {
		cmd := exec.CommandContext(ctx, pem.uvPath, "pip", "install", "openpyxl", "--python", pythonBin)
		hideWindow(cmd)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install dependencies: %w\n%s", err, string(output))
		}
	}

	return nil
}

// RunScript starts a Python script in the managed virtual environment and returns
// the command, stdout pipe, and stderr pipe for the caller to read from.
func (pem *PythonEnvManager) RunScript(ctx context.Context, scriptPath string, args []string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	pythonBin := pem.pythonPath()

	cmdArgs := make([]string, 0, 1+len(args))
	cmdArgs = append(cmdArgs, scriptPath)
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.CommandContext(ctx, pythonBin, cmdArgs...)
	hideWindow(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		return nil, nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return nil, nil, nil, fmt.Errorf("failed to start python script: %w", err)
	}

	return cmd, stdout, stderr, nil
}

// GetStatus returns the current state of the Python environment, including
// whether uv is available and whether the virtual environment exists.
func (pem *PythonEnvManager) GetStatus() *EnvStatus {
	uvAvailable := pem.checkUv(context.Background()) == nil
	_, err := os.Stat(pem.envPath)
	envExists := err == nil

	return &EnvStatus{
		UvAvailable: uvAvailable,
		EnvExists:   envExists,
		EnvPath:     pem.envPath,
	}
}

// checkUv verifies that the uv tool is available by running `uv --version`.
func (pem *PythonEnvManager) checkUv(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, pem.uvPath, "--version")
	hideWindow(cmd)
	return cmd.Run()
}

// PythonPath returns the public path to the Python binary inside the virtual environment.
func (pem *PythonEnvManager) PythonPath() string {
	return pem.pythonPath()
}

// pythonPath returns the path to the Python binary inside the virtual environment.
// On Windows it uses Scripts/python.exe, on other platforms it uses bin/python.
func (pem *PythonEnvManager) pythonPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(pem.envPath, "Scripts", "python.exe")
	}
	return filepath.Join(pem.envPath, "bin", "python")
}

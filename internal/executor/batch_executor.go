package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"network-log-formatter/internal/model"
	"network-log-formatter/internal/pyenv"
)

// LLMRepairer defines the interface for LLM-based code repair.
// This avoids circular imports with the agent package.
type LLMRepairer interface {
	RepairCode(ctx context.Context, code string, errorMsg string) (string, error)
}

// BatchExecutor runs generated Python scripts in a uv-managed environment,
// monitors progress via stdout, and handles runtime error auto-repair.
type BatchExecutor struct {
	envManager *pyenv.PythonEnvManager
	llmClient  LLMRepairer
	maxRetries int
	progress   *model.BatchProgress
	mu         sync.Mutex
}

// NewBatchExecutor creates a new BatchExecutor with the given dependencies.
func NewBatchExecutor(envManager *pyenv.PythonEnvManager, llmClient LLMRepairer, maxRetries int) *BatchExecutor {
	return &BatchExecutor{
		envManager: envManager,
		llmClient:  llmClient,
		maxRetries: maxRetries,
		progress: &model.BatchProgress{
			Status: "idle",
		},
	}
}

// Execute runs the given Python code against the input directory and writes
// results to the output directory. It monitors stdout for JSON progress lines
// and stderr for errors. If a runtime error occurs, it sends the code and error
// to the LLM for repair and retries up to maxRetries times.
func (be *BatchExecutor) Execute(ctx context.Context, code string, inputDir string, outputDir string) (*model.BatchResult, error) {
	// Validate directories
	if strings.TrimSpace(inputDir) == "" {
		return nil, fmt.Errorf("input directory must not be empty")
	}
	if strings.TrimSpace(outputDir) == "" {
		return nil, fmt.Errorf("output directory must not be empty")
	}

	// Validate inputDir exists
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("input directory does not exist: %s", inputDir)
	}

	// Validate paths are absolute to prevent traversal issues
	absInput, err := filepath.Abs(inputDir)
	if err != nil {
		return nil, fmt.Errorf("invalid input directory path: %w", err)
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("invalid output directory path: %w", err)
	}
	inputDir = absInput
	outputDir = absOutput

	// Create outputDir if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	currentCode := code
	var lastErr string

	for attempt := 0; attempt <= be.maxRetries; attempt++ {
		result, stderrOutput, err := be.runScript(ctx, currentCode, inputDir, outputDir)
		if err == nil {
			// Process exited successfully (exit code 0).
			// stderr may contain informational messages — that's fine.
			be.setProgress(&model.BatchProgress{
				Status:     "completed",
				TotalFiles: result.TotalFiles,
				Processed:  result.Succeeded,
				Failed:     result.Failed,
				Progress:   1.0,
				Message:    "Batch processing completed",
			})
			return result, nil
		}

		// Determine error message
		if err != nil {
			if stderrOutput != "" {
				lastErr = stderrOutput
			} else {
				lastErr = err.Error()
			}
		}

		// If we've exhausted retries, return failure
		if attempt >= be.maxRetries {
			break
		}

		// Try LLM repair
		be.setProgress(&model.BatchProgress{
			Status:  "fixing",
			Message: fmt.Sprintf("Runtime error detected, attempting repair (attempt %d/%d)", attempt+1, be.maxRetries),
		})

		if be.llmClient == nil {
			break
		}

		fixedCode, repairErr := be.llmClient.RepairCode(ctx, currentCode, lastErr)
		if repairErr != nil {
			// Can't repair, return the original error
			break
		}
		currentCode = fixedCode
	}

	// Failed after all retries
	be.setProgress(&model.BatchProgress{
		Status:  "failed",
		Message: fmt.Sprintf("Batch processing failed: %s", lastErr),
	})

	return &model.BatchResult{
		Errors: []string{lastErr},
	}, fmt.Errorf("batch execution failed after %d retries: %s", be.maxRetries, lastErr)
}

// runScript writes the code to a temp file, executes it via PythonEnvManager,
// and reads stdout/stderr concurrently. Returns the batch result and any stderr output.
func (be *BatchExecutor) runScript(ctx context.Context, code string, inputDir string, outputDir string) (*model.BatchResult, string, error) {
	// Write code to temp file
	tmpDir, err := os.MkdirTemp("", "batch-executor-*")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "script.py")
	if err := os.WriteFile(scriptPath, []byte(code), 0644); err != nil {
		return nil, "", fmt.Errorf("failed to write temp script: %w", err)
	}

	be.setProgress(&model.BatchProgress{
		Status:  "running",
		Message: "Starting batch processing",
	})

	// Run script with --input and --output args
	args := []string{"--input", inputDir, "--output", outputDir}
	cmd, stdout, stderr, err := be.envManager.RunScript(ctx, scriptPath, args)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start script: %w", err)
	}

	// Read stdout and stderr concurrently
	var stderrBuf strings.Builder
	result := &model.BatchResult{
		OutputPath: outputDir,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Read stdout — parse JSON progress lines
	go func() {
		defer wg.Done()
		be.readStdout(stdout, result)
	}()

	// Read stderr
	go func() {
		defer wg.Done()
		data, _ := io.ReadAll(stderr)
		stderrBuf.WriteString(string(data))
	}()

	wg.Wait()

	// Wait for the process to finish
	waitErr := cmd.Wait()

	stderrOutput := strings.TrimSpace(stderrBuf.String())

	if waitErr != nil {
		if stderrOutput != "" {
			return result, stderrOutput, waitErr
		}
		return result, "", waitErr
	}

	return result, stderrOutput, nil
}

// readStdout reads stdout line by line, parsing JSON progress lines and updating
// the batch progress and result accordingly.
func (be *BatchExecutor) readStdout(stdout io.ReadCloser, result *model.BatchResult) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var info model.ProgressInfo
		if err := json.Unmarshal([]byte(line), &info); err != nil {
			// Not a JSON progress line, skip
			continue
		}

		// Update progress
		be.setProgress(&model.BatchProgress{
			Status:      "running",
			CurrentFile: info.File,
			Progress:    info.Progress,
			TotalFiles:  info.Total,
			Processed:   info.Current,
			Message:     fmt.Sprintf("Processing: %s", info.File),
		})

		// Update result counts from the latest progress info (protected by mutex)
		be.mu.Lock()
		result.TotalFiles = info.Total
		result.Succeeded = info.Current
		be.mu.Unlock()
	}
}

// GetProgress returns the current batch processing progress (thread-safe).
func (be *BatchExecutor) GetProgress() *model.BatchProgress {
	be.mu.Lock()
	defer be.mu.Unlock()

	// Return a copy to avoid data races
	p := *be.progress
	return &p
}

// setProgress updates the current progress (thread-safe).
func (be *BatchExecutor) setProgress(p *model.BatchProgress) {
	be.mu.Lock()
	defer be.mu.Unlock()
	be.progress = p
}

package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"network-log-formatter/internal/model"
	"network-log-formatter/internal/pyenv"
)

// ValidationResult holds the outcome of a code validation attempt.
type ValidationResult struct {
	Valid   bool
	Code    string
	Errors  []string
	Retries int
}

// CodeValidator validates generated Python code via syntax checking
// and triggers LLM-based auto-repair on failure.
type CodeValidator struct {
	envManager *pyenv.PythonEnvManager
	llmClient  *LLMClient
	maxRetries int
}

// NewCodeValidator creates a new CodeValidator.
func NewCodeValidator(envManager *pyenv.PythonEnvManager, llmClient *LLMClient, maxRetries int) *CodeValidator {
	return &CodeValidator{
		envManager: envManager,
		llmClient:  llmClient,
		maxRetries: maxRetries,
	}
}

// Validate checks the given Python code for syntax errors. If the syntax check
// fails, it sends the code and error to the LLM for repair and retries up to
// maxRetries times. Temp files are cleaned up after validation.
func (cv *CodeValidator) Validate(ctx context.Context, code string) (*ValidationResult, error) {
	result := &ValidationResult{
		Code:   code,
		Errors: []string{},
	}

	currentCode := code

	for attempt := 0; attempt <= cv.maxRetries; attempt++ {
		syntaxErr, err := cv.checkSyntax(ctx, currentCode)
		if err != nil {
			return nil, fmt.Errorf("syntax check execution failed: %w", err)
		}

		if syntaxErr == "" {
			// Syntax check passed
			result.Valid = true
			result.Code = currentCode
			result.Retries = attempt
			return result, nil
		}

		// Syntax check failed — collect the error
		result.Errors = append(result.Errors, syntaxErr)

		// If we've exhausted retries, stop
		if attempt >= cv.maxRetries {
			break
		}

		// Ask LLM to repair the code
		fixedCode, err := cv.repairCode(ctx, currentCode, syntaxErr)
		if err != nil {
			return nil, fmt.Errorf("LLM repair failed: %w", err)
		}

		currentCode = fixedCode
	}

	// Max retries reached without passing validation
	result.Valid = false
	result.Code = currentCode
	result.Retries = cv.maxRetries
	return result, nil
}

// checkSyntax writes the code to a temp file and runs the Python compiler on it.
// Returns the error message from stderr if syntax check fails, or empty string if OK.
func (cv *CodeValidator) checkSyntax(ctx context.Context, code string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "code-validator-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "temp.py")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	pythonBin := cv.envManager.PythonPath()
	compileCmd := fmt.Sprintf("compile(open(%q).read(), %q, 'exec')", tmpFile, tmpFile)
	cmd := exec.CommandContext(ctx, pythonBin, "-c", compileCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Syntax check failed — return the error output
		errMsg := strings.TrimSpace(string(output))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return errMsg, nil
	}

	return "", nil
}

// repairCode sends the code and its syntax error to the LLM for repair,
// then extracts the fixed Python code from the response.
func (cv *CodeValidator) repairCode(ctx context.Context, code string, syntaxErr string) (string, error) {
	messages := []model.Message{
		{
			Role: "system",
			Content: "You are an expert Python developer. Fix the syntax error in the given Python code. " +
				"Return the complete fixed Python code inside a single ```python code block. " +
				"Do not explain the changes, just return the corrected code.",
		},
		{
			Role: "user",
			Content: fmt.Sprintf("The following Python code has a syntax error:\n\n```python\n%s\n```\n\n"+
				"Error message:\n```\n%s\n```\n\nPlease fix the syntax error and return the complete corrected code.",
				code, syntaxErr),
		},
	}

	resp, err := cv.llmClient.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	return extractCode(resp), nil
}

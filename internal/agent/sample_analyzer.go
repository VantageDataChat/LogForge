package agent

import (
	"context"
	"errors"
	"strings"

	"network-log-formatter/internal/model"
)

// SampleAnalyzer sends sample log entries to LLM and retrieves generated Python code.
type SampleAnalyzer struct {
	llmClient *LLMClient
}

// NewSampleAnalyzer creates a new SampleAnalyzer with the given LLM client.
func NewSampleAnalyzer(llmClient *LLMClient) *SampleAnalyzer {
	return &SampleAnalyzer{llmClient: llmClient}
}

// Analyze validates the sample input, builds a prompt, calls the LLM, and extracts
// the generated Python code from the response.
func (sa *SampleAnalyzer) Analyze(ctx context.Context, sampleText string) (string, error) {
	if strings.TrimSpace(sampleText) == "" {
		return "", errors.New("sample text must not be empty")
	}

	messages := []model.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: buildUserPrompt(sampleText),
		},
	}

	resp, err := sa.llmClient.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	return extractCode(resp), nil
}

const systemPrompt = `You are an expert Python developer specializing in log parsing and data processing.
Your task is to analyze sample log entries and generate a complete Python program that can batch-process log files of the same format.

The generated Python program MUST:
1. Accept --input and --output command line arguments (--input is the directory containing log files, --output is the directory for Excel output)
2. Traverse all log files in the input directory
3. Parse each log entry into structured data based on the detected format
4. Use openpyxl to merge ALL parsed data from ALL log files into a SINGLE Excel file (e.g. output/result.xlsx). Do NOT create one Excel per log file. All rows go into one worksheet so the user can process them together easily.
5. Output progress to stdout as JSON lines, one per file processed, in this exact format:
   {"file": "<filename>", "progress": <0.0-1.0>, "total": <total_files>, "current": <current_index>}
6. Include complete error handling (try/except around file operations, graceful handling of unparseable entries)
7. Add a "source_file" column so the user knows which log file each row came from

Return the complete Python code inside a single python code block.`

func buildUserPrompt(sampleText string) string {
	return "Please analyze the following sample log entries and generate a complete Python processing program.\n\n" +
		"Sample log entries:\n```\n" + sampleText + "\n```"
}

// extractCode extracts Python code from an LLM response.
// It first looks for a ```python ... ``` block, then any ``` ... ``` block.
// Returns an error-indicating empty string if no code block is found.
func extractCode(response string) string {
	// Try to find ```python ... ``` block
	if code, ok := extractFencedBlock(response, "```python"); ok {
		return code
	}

	// Try to find any ``` ... ``` block
	if code, ok := extractFencedBlock(response, "```"); ok {
		return code
	}

	// No code block found — return trimmed response only if it looks like Python code
	// (contains common Python keywords), otherwise return as-is and let the validator catch it
	trimmed := strings.TrimSpace(response)
	if strings.Contains(trimmed, "import ") || strings.Contains(trimmed, "def ") || strings.Contains(trimmed, "class ") {
		return trimmed
	}

	return trimmed
}

// extractFencedBlock finds the first occurrence of a fenced code block starting
// with the given prefix and returns its content.
func extractFencedBlock(text, prefix string) (string, bool) {
	start := strings.Index(text, prefix)
	if start == -1 {
		return "", false
	}

	// Move past the prefix line
	contentStart := strings.Index(text[start:], "\n")
	if contentStart == -1 {
		return "", false
	}
	contentStart += start + 1

	// Find the closing ```
	end := strings.Index(text[contentStart:], "```")
	if end == -1 {
		return "", false
	}

	code := text[contentStart : contentStart+end]
	return strings.TrimSpace(code), true
}

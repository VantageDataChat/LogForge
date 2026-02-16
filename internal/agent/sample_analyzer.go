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

	code := extractCode(resp)
	if code == "" {
		return "", errors.New("LLM response did not contain valid Python code")
	}

	return code, nil
}

const systemPrompt = `You are an expert Python developer specializing in log parsing and data processing.
Your task is to analyze sample log entries and generate a complete Python program that can batch-process log files of the same format.

The generated Python program MUST:
1. Accept --input and --output command line arguments (--input is the directory containing log files, --output is the directory for Excel output)
2. Traverse all log files in the input directory
3. Parse each log entry into structured data based on the detected format
4. Use openpyxl to write ALL parsed data into a SINGLE Excel file (e.g. output/result.xlsx), but create a SEPARATE SHEET for each input log file. The sheet name MUST be the original log file name WITH extension (e.g. "Apache_2k.log"). If the file name exceeds 31 characters (Excel sheet name limit), truncate it to 31 characters. Do NOT use generic names like "Log Entries" or "Sheet1". Do NOT merge all data into one worksheet.
   IMPORTANT: Each log file must produce exactly ONE sheet. Do NOT create duplicate sheets. When creating the Workbook, immediately remove the default empty sheet (wb.remove(wb.active)) before adding any data sheets. Ensure each file is only processed once.
5. STRICTLY FORBIDDEN extra columns:
   - Do NOT add a "source_file" column. The sheet name already identifies the source file.
   - Do NOT add a row number / line number / index / sequence column.
   - Do NOT add a "raw_log" / "raw_line" / "original" / "raw" column containing the original log line text.
   - The Excel output must ONLY contain the parsed/structured data fields (e.g. datetime, level, module, pid, message). No redundant or auxiliary columns.
6. For date/time fields: if the log contains date and time information that appears on multiple lines (e.g. a date header followed by time-only entries), consolidate them so each row has ONE complete datetime or date column. Do NOT repeat the same date across a separate column. Keep only one unified date/time column per row to make statistical analysis easier.
7. Output progress to stdout as JSON lines, one per file processed, in this exact format:
   {"file": "<filename>", "progress": <0.0-1.0>, "total": <total_files>, "current": <current_index>}
8. Include complete error handling (try/except around file operations, graceful handling of unparseable entries)

Return the complete Python code inside a single python code block.`

func buildUserPrompt(sampleText string) string {
	return "Please analyze the following sample log entries and generate a complete Python processing program.\n\n" +
		"Sample log entries:\n```\n" + sampleText + "\n```"
}

// extractCode extracts Python code from an LLM response.
// It first looks for a ```python ... ``` block, then any ``` ... ``` block.
// Falls back to the raw response only if it looks like Python code.
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
	trimmed := strings.TrimSpace(response)
	if strings.Contains(trimmed, "import ") || strings.Contains(trimmed, "def ") ||
		strings.Contains(trimmed, "class ") || strings.Contains(trimmed, "print(") ||
		strings.Contains(trimmed, "if __name__") {
		return trimmed
	}

	// Doesn't look like Python code at all — return empty string so the caller
	// can detect the failure via validation rather than executing arbitrary text
	return ""
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

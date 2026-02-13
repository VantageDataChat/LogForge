package executor

import (
	"context"
	"os"
	"testing"

	"pgregory.net/rapid"
)

// Feature: network-log-formatter, Property 5: 批处理目录必填验证
// For any batch processing request, if input directory or output directory is
// empty (or whitespace-only), the system should reject execution and return an error.
// **Validates: Requirements 4.1, 4.6**
func TestProperty5_BatchDirectoryRequiredValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random non-empty directory string for the "valid" side
		validDir := rapid.StringMatching(`[a-zA-Z0-9_/\\]{1,50}`).Draw(t, "validDir")

		// Decide which directory to make empty/whitespace:
		// 0 = both empty, 1 = inputDir empty, 2 = outputDir empty
		mode := rapid.IntRange(0, 2).Draw(t, "mode")

		// Generate an empty or whitespace-only string
		emptyGen := rapid.SampledFrom([]string{"", " ", "  ", "\t", "\n", " \t\n "})
		emptyVal := emptyGen.Draw(t, "emptyVal")

		var inputDir, outputDir string
		switch mode {
		case 0:
			inputDir = emptyGen.Draw(t, "emptyInput")
			outputDir = emptyGen.Draw(t, "emptyOutput")
		case 1:
			inputDir = emptyVal
			outputDir = validDir
		case 2:
			inputDir = validDir
			outputDir = emptyVal
		}

		// Create executor with nil dependencies — validation should fail before they're used
		be := NewBatchExecutor(nil, nil, 3)

		_, err := be.Execute(context.Background(), "print('hello')", inputDir, outputDir)

		// Property: must return an error when at least one directory is empty/whitespace
		if err == nil {
			t.Fatalf("expected error for inputDir=%q outputDir=%q (mode=%d), got nil",
				inputDir, outputDir, mode)
		}
	})
}

// --- Unit Tests ---

// Unit test: both directories empty
func TestExecute_BothDirsEmpty(t *testing.T) {
	be := NewBatchExecutor(nil, nil, 3)
	_, err := be.Execute(context.Background(), "print('hello')", "", "")
	if err == nil {
		t.Fatal("expected error when both directories are empty")
	}
}

// Unit test: only inputDir empty
func TestExecute_InputDirEmpty(t *testing.T) {
	be := NewBatchExecutor(nil, nil, 3)
	_, err := be.Execute(context.Background(), "print('hello')", "", "/tmp/output")
	if err == nil {
		t.Fatal("expected error when inputDir is empty")
	}
}

// Unit test: only outputDir empty
func TestExecute_OutputDirEmpty(t *testing.T) {
	be := NewBatchExecutor(nil, nil, 3)

	// Use a real existing directory for inputDir so we pass that validation
	tmpDir, err := os.MkdirTemp("", "batch-test-input-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, execErr := be.Execute(context.Background(), "print('hello')", tmpDir, "")
	if execErr == nil {
		t.Fatal("expected error when outputDir is empty")
	}
}

// Unit test: non-existent inputDir
func TestExecute_NonExistentInputDir(t *testing.T) {
	be := NewBatchExecutor(nil, nil, 3)
	_, err := be.Execute(context.Background(), "print('hello')", "/nonexistent/path/abc123", "/tmp/output")
	if err == nil {
		t.Fatal("expected error when inputDir does not exist")
	}
}

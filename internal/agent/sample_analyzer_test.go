package agent

import (
	"context"
	"testing"

	"pgregory.net/rapid"
)

// Feature: network-log-formatter, Property 1: 空样本输入拒绝
// For any empty or whitespace-only string as sample input, AnalyzeSample should
// reject the request and return an error without calling the LLM.
// **Validates: Requirements 1.4**
func TestProperty1_EmptySampleInputRejection(t *testing.T) {
	// Use a nil LLMClient: if Analyze ever tries to call the LLM,
	// it will panic on nil pointer dereference, proving the LLM was called.
	sa := NewSampleAnalyzer(nil)

	rapid.Check(t, func(t *rapid.T) {
		// Generate random whitespace-only strings from spaces, tabs, newlines
		whitespaceChars := []rune{' ', '\t', '\n', '\r'}
		length := rapid.IntRange(0, 50).Draw(t, "length")
		runes := make([]rune, length)
		for i := range runes {
			runes[i] = whitespaceChars[rapid.IntRange(0, len(whitespaceChars)-1).Draw(t, "charIdx")]
		}
		input := string(runes)

		_, err := sa.Analyze(context.Background(), input)
		if err == nil {
			t.Fatalf("expected error for whitespace-only input %q, got nil", input)
		}
	})
}

// Unit tests for extractCode function

func TestExtractCode_PythonBlock(t *testing.T) {
	response := "Here is the code:\n```python\nprint('hello')\n```\nDone."
	got := extractCode(response)
	want := "print('hello')"
	if got != want {
		t.Errorf("extractCode with python block: got %q, want %q", got, want)
	}
}

func TestExtractCode_GenericBlock(t *testing.T) {
	response := "Here is the code:\n```\nimport os\nos.listdir('.')\n```\nDone."
	got := extractCode(response)
	want := "import os\nos.listdir('.')"
	if got != want {
		t.Errorf("extractCode with generic block: got %q, want %q", got, want)
	}
}

func TestExtractCode_NoBlock(t *testing.T) {
	response := "print('hello world')"
	got := extractCode(response)
	if got != response {
		t.Errorf("extractCode with no block: got %q, want %q", got, response)
	}
}

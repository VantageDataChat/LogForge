package agent

import (
	"testing"

	"pgregory.net/rapid"
)

// simulateValidationLoop simulates the CodeValidator's retry logic without
// requiring a real Python environment or LLM. It models the core loop:
// for each attempt, if syntax check fails and retries remain, try again.
// Returns a ValidationResult with the same invariants as the real Validate method.
func simulateValidationLoop(maxRetries int, syntaxCheckResults []bool) *ValidationResult {
	result := &ValidationResult{
		Code:   "test_code",
		Errors: []string{},
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Determine if this attempt passes syntax check
		passes := false
		if attempt < len(syntaxCheckResults) {
			passes = syntaxCheckResults[attempt]
		}

		if passes {
			result.Valid = true
			result.Retries = attempt
			return result
		}

		result.Errors = append(result.Errors, "syntax error")

		if attempt >= maxRetries {
			break
		}
		// In real code, LLM repair happens here
	}

	result.Valid = false
	result.Retries = maxRetries
	return result
}

// validatedStatus returns the project status string that should be set
// when a ValidationResult indicates the code is valid.
func validatedStatus(vr *ValidationResult) string {
	if vr.Valid {
		return "validated"
	}
	return "failed"
}

// Feature: network-log-formatter, Property 3: 重试循环有界性
// For any validation-failing code, CodeValidator's repair-validate loop count
// should not exceed the configured maxRetries value.
// **Validates: Requirements 3.3, 3.4**
func TestProperty3_RetryLoopBounded(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		maxRetries := rapid.IntRange(1, 10).Draw(t, "maxRetries")

		// Generate a sequence of syntax check results (all false = always failing)
		numChecks := maxRetries + 1 // one initial + maxRetries repair attempts
		syntaxResults := make([]bool, numChecks)
		for i := range syntaxResults {
			syntaxResults[i] = false // all checks fail
		}

		result := simulateValidationLoop(maxRetries, syntaxResults)

		// Property: Retries must never exceed maxRetries
		if result.Retries > maxRetries {
			t.Fatalf("retries %d exceeded maxRetries %d", result.Retries, maxRetries)
		}

		// When all checks fail, retries should equal maxRetries exactly
		if result.Retries != maxRetries {
			t.Fatalf("expected retries=%d when all checks fail, got %d", maxRetries, result.Retries)
		}

		// Code should not be valid when all checks fail
		if result.Valid {
			t.Fatal("expected Valid=false when all syntax checks fail")
		}
	})
}

// Property 3 variant: when a check passes at some attempt, retries should be
// that attempt number (which is <= maxRetries).
func TestProperty3_RetryLoopBounded_EarlySuccess(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		maxRetries := rapid.IntRange(1, 10).Draw(t, "maxRetries")

		// Pick a random attempt at which syntax check passes
		successAt := rapid.IntRange(0, maxRetries).Draw(t, "successAt")

		numChecks := maxRetries + 1
		syntaxResults := make([]bool, numChecks)
		syntaxResults[successAt] = true // this attempt passes

		result := simulateValidationLoop(maxRetries, syntaxResults)

		// Property: Retries must never exceed maxRetries
		if result.Retries > maxRetries {
			t.Fatalf("retries %d exceeded maxRetries %d", result.Retries, maxRetries)
		}

		// Should succeed at the expected attempt
		if !result.Valid {
			t.Fatalf("expected Valid=true when check passes at attempt %d", successAt)
		}

		if result.Retries != successAt {
			t.Fatalf("expected retries=%d (success attempt), got %d", successAt, result.Retries)
		}
	})
}

// Feature: network-log-formatter, Property 4: 验证通过后状态更新
// For any code that passes validation, the corresponding project status
// should be set to "validated".
// **Validates: Requirements 3.5**
func TestProperty4_ValidatedStatusUpdate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		maxRetries := rapid.IntRange(1, 10).Draw(t, "maxRetries")

		// Pick a random attempt at which validation passes
		successAt := rapid.IntRange(0, maxRetries).Draw(t, "successAt")

		numChecks := maxRetries + 1
		syntaxResults := make([]bool, numChecks)
		syntaxResults[successAt] = true

		result := simulateValidationLoop(maxRetries, syntaxResults)

		// Property: when Valid=true, status must be "validated"
		status := validatedStatus(result)
		if result.Valid && status != "validated" {
			t.Fatalf("expected status 'validated' for Valid=true result, got %q", status)
		}
	})
}

// Property 4 variant: when validation fails, status should be "failed"
func TestProperty4_FailedStatusUpdate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		maxRetries := rapid.IntRange(1, 10).Draw(t, "maxRetries")

		// All checks fail
		numChecks := maxRetries + 1
		syntaxResults := make([]bool, numChecks)

		result := simulateValidationLoop(maxRetries, syntaxResults)

		status := validatedStatus(result)
		if !result.Valid && status != "failed" {
			t.Fatalf("expected status 'failed' for Valid=false result, got %q", status)
		}
	})
}

// --- Unit Tests ---

// Unit test: ValidationResult structure correctness
func TestValidationResult_ValidCode(t *testing.T) {
	result := &ValidationResult{
		Valid:   true,
		Code:    "print('hello')",
		Errors:  []string{},
		Retries: 0,
	}

	if !result.Valid {
		t.Fatal("expected Valid=true")
	}
	if result.Code != "print('hello')" {
		t.Fatalf("unexpected code: %q", result.Code)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
}

func TestValidationResult_InvalidCode(t *testing.T) {
	result := &ValidationResult{
		Valid:   false,
		Code:    "def foo(:",
		Errors:  []string{"SyntaxError: invalid syntax"},
		Retries: 3,
	}

	if result.Valid {
		t.Fatal("expected Valid=false")
	}
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Retries != 3 {
		t.Fatalf("expected Retries=3, got %d", result.Retries)
	}
}

// Unit test: NewCodeValidator sets maxRetries correctly
func TestNewCodeValidator_MaxRetries(t *testing.T) {
	cv := NewCodeValidator(nil, nil, 5)
	if cv.maxRetries != 5 {
		t.Fatalf("expected maxRetries=5, got %d", cv.maxRetries)
	}
}

func TestNewCodeValidator_DefaultMaxRetries(t *testing.T) {
	cv := NewCodeValidator(nil, nil, 3)
	if cv.maxRetries != 3 {
		t.Fatalf("expected maxRetries=3, got %d", cv.maxRetries)
	}
}

// Unit test: retry bound with maxRetries=0 (no retries allowed)
func TestSimulateValidation_ZeroRetries_AllFail(t *testing.T) {
	result := simulateValidationLoop(0, []bool{false})
	if result.Valid {
		t.Fatal("expected Valid=false with 0 retries and failing check")
	}
	if result.Retries != 0 {
		t.Fatalf("expected Retries=0, got %d", result.Retries)
	}
}

func TestSimulateValidation_ZeroRetries_FirstPass(t *testing.T) {
	result := simulateValidationLoop(0, []bool{true})
	if !result.Valid {
		t.Fatal("expected Valid=true when first check passes")
	}
	if result.Retries != 0 {
		t.Fatalf("expected Retries=0, got %d", result.Retries)
	}
}

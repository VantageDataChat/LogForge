package model

import (
	"encoding/json"
	"testing"

	"pgregory.net/rapid"
)

// Feature: network-log-formatter, Property 6: 进度信息 JSON 解析正确性
// For any valid ProgressInfo struct, serializing to JSON then deserializing
// should produce an equivalent result.
// **Validates: Requirements 4.2**
func TestProperty6_ProgressInfoJSONRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		original := ProgressInfo{
			File:     rapid.String().Draw(t, "file"),
			Progress: rapid.Float64Range(0, 1).Draw(t, "progress"),
			Total:    rapid.IntRange(0, 100000).Draw(t, "total"),
			Current:  rapid.IntRange(0, 100000).Draw(t, "current"),
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal ProgressInfo: %v", err)
		}

		var decoded ProgressInfo
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("failed to unmarshal ProgressInfo: %v", err)
		}

		if original.File != decoded.File {
			t.Fatalf("File mismatch: got %q, want %q", decoded.File, original.File)
		}
		if original.Progress != decoded.Progress {
			t.Fatalf("Progress mismatch: got %v, want %v", decoded.Progress, original.Progress)
		}
		if original.Total != decoded.Total {
			t.Fatalf("Total mismatch: got %d, want %d", decoded.Total, original.Total)
		}
		if original.Current != decoded.Current {
			t.Fatalf("Current mismatch: got %d, want %d", decoded.Current, original.Current)
		}
	})
}

// Feature: network-log-formatter, Property 7: 批处理结果不变量
// For any BatchResult, TotalFiles should equal Succeeded + Failed.
// **Validates: Requirements 4.5**
func TestProperty7_BatchResultInvariant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		succeeded := rapid.IntRange(0, 50000).Draw(t, "succeeded")
		failed := rapid.IntRange(0, 50000).Draw(t, "failed")

		result := BatchResult{
			TotalFiles: succeeded + failed,
			Succeeded:  succeeded,
			Failed:     failed,
			OutputPath: rapid.String().Draw(t, "outputPath"),
		}

		if result.TotalFiles != result.Succeeded+result.Failed {
			t.Fatalf("invariant violated: TotalFiles(%d) != Succeeded(%d) + Failed(%d)",
				result.TotalFiles, result.Succeeded, result.Failed)
		}
	})
}

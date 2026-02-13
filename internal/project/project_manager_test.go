package project

import (
	"os"
	"testing"
	"time"

	"network-log-formatter/internal/model"

	"pgregory.net/rapid"
)

// alphanumericID generates a non-empty alphanumeric string safe for filenames.
func alphanumericID() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-zA-Z0-9]{1,32}`)
}

// Feature: network-log-formatter, Property 8: 项目创建完整性
// For any successful code generation, the created project should contain all required fields:
// non-empty ID, sample data, code, creation time, status.
// **Validates: Requirements 5.1, 5.2**
func TestProperty8_ProjectCreationCompleteness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tmpDir, err := os.MkdirTemp("", "project-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		pm, err := NewProjectManager(tmpDir)
		if err != nil {
			t.Fatalf("failed to create ProjectManager: %v", err)
		}

		id := alphanumericID().Draw(t, "id")
		sampleData := rapid.String().Draw(t, "sampleData")
		code := rapid.String().Draw(t, "code")
		status := rapid.SampledFrom([]string{"draft", "validated", "executed", "failed"}).Draw(t, "status")
		now := time.Now()

		project := model.Project{
			ID:         id,
			SampleData: sampleData,
			Code:       code,
			CreatedAt:  now,
			UpdatedAt:  now,
			Status:     status,
		}

		if err := pm.Create(project); err != nil {
			t.Fatalf("failed to create project: %v", err)
		}

		got, err := pm.Get(id)
		if err != nil {
			t.Fatalf("failed to get project: %v", err)
		}

		if got.ID == "" {
			t.Fatal("project ID should not be empty")
		}
		if got.ID != id {
			t.Fatalf("ID mismatch: got %q, want %q", got.ID, id)
		}
		if got.SampleData != sampleData {
			t.Fatalf("SampleData mismatch: got %q, want %q", got.SampleData, sampleData)
		}
		if got.Code != code {
			t.Fatalf("Code mismatch: got %q, want %q", got.Code, code)
		}
		if got.CreatedAt.IsZero() {
			t.Fatal("CreatedAt should not be zero")
		}
		if got.Status == "" {
			t.Fatal("Status should not be empty")
		}
		if got.Status != status {
			t.Fatalf("Status mismatch: got %q, want %q", got.Status, status)
		}
	})
}

// Feature: network-log-formatter, Property 9: 项目列表排序
// For any project list query, projects should be sorted by creation time descending.
// **Validates: Requirements 5.3**
func TestProperty9_ProjectListSorting(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tmpDir, err := os.MkdirTemp("", "project-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		pm, err := NewProjectManager(tmpDir)
		if err != nil {
			t.Fatalf("failed to create ProjectManager: %v", err)
		}

		// Generate between 2 and 10 projects with distinct timestamps.
		count := rapid.IntRange(2, 10).Draw(t, "count")
		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		for i := 0; i < count; i++ {
			// Use random offsets to create varied timestamps.
			offsetMinutes := rapid.IntRange(0, 525600).Draw(t, "offsetMinutes") // up to ~1 year
			project := model.Project{
				ID:         alphanumericID().Draw(t, "id"),
				SampleData: "sample",
				Code:       "code",
				CreatedAt:  baseTime.Add(time.Duration(offsetMinutes) * time.Minute),
				UpdatedAt:  baseTime.Add(time.Duration(offsetMinutes) * time.Minute),
				Status:     "draft",
			}
			if err := pm.Create(project); err != nil {
				t.Fatalf("failed to create project: %v", err)
			}
		}

		projects, err := pm.List()
		if err != nil {
			t.Fatalf("failed to list projects: %v", err)
		}

		// Verify descending order by CreatedAt.
		for i := 1; i < len(projects); i++ {
			if projects[i-1].CreatedAt.Before(projects[i].CreatedAt) {
				t.Fatalf("projects not sorted descending: index %d (%v) is before index %d (%v)",
					i-1, projects[i-1].CreatedAt, i, projects[i].CreatedAt)
			}
		}
	})
}

// Feature: network-log-formatter, Property 10: 项目代码更新往返
// For any project and new code string, updating then getting should return the new code.
// **Validates: Requirements 5.4**
func TestProperty10_ProjectCodeUpdateRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tmpDir, err := os.MkdirTemp("", "project-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		pm, err := NewProjectManager(tmpDir)
		if err != nil {
			t.Fatalf("failed to create ProjectManager: %v", err)
		}

		id := alphanumericID().Draw(t, "id")
		now := time.Now()
		project := model.Project{
			ID:         id,
			SampleData: "sample",
			Code:       "original code",
			CreatedAt:  now,
			UpdatedAt:  now,
			Status:     "draft",
		}
		if err := pm.Create(project); err != nil {
			t.Fatalf("failed to create project: %v", err)
		}

		newCode := rapid.String().Draw(t, "newCode")
		update := model.ProjectUpdate{
			Code: &newCode,
		}
		if err := pm.Update(id, update); err != nil {
			t.Fatalf("failed to update project: %v", err)
		}

		got, err := pm.Get(id)
		if err != nil {
			t.Fatalf("failed to get project: %v", err)
		}

		if got.Code != newCode {
			t.Fatalf("Code mismatch after update: got %q, want %q", got.Code, newCode)
		}
	})
}

// Feature: network-log-formatter, Property 11: 项目删除有效性
// For any existing project, after deletion it should not appear in the list.
// **Validates: Requirements 5.5**
func TestProperty11_ProjectDeletionValidity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tmpDir, err := os.MkdirTemp("", "project-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		pm, err := NewProjectManager(tmpDir)
		if err != nil {
			t.Fatalf("failed to create ProjectManager: %v", err)
		}

		id := alphanumericID().Draw(t, "id")
		now := time.Now()
		project := model.Project{
			ID:         id,
			SampleData: "sample",
			Code:       "code",
			CreatedAt:  now,
			UpdatedAt:  now,
			Status:     "draft",
		}
		if err := pm.Create(project); err != nil {
			t.Fatalf("failed to create project: %v", err)
		}

		if err := pm.Delete(id); err != nil {
			t.Fatalf("failed to delete project: %v", err)
		}

		projects, err := pm.List()
		if err != nil {
			t.Fatalf("failed to list projects: %v", err)
		}

		for _, p := range projects {
			if p.ID == id {
				t.Fatalf("deleted project %q still appears in list", id)
			}
		}

		// Also verify Get returns an error.
		_, err = pm.Get(id)
		if err == nil {
			t.Fatalf("expected error when getting deleted project %q, got nil", id)
		}
	})
}

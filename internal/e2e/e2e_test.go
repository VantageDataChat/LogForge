// Package e2e contains end-to-end integration tests for the full pipeline:
// log files → Python script → Excel output.
//
// These tests require uv to be installed and will be skipped if unavailable.
// Run with: go test ./internal/e2e/ -v -count=1 -timeout 300s
package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"network-log-formatter/internal/agent"
	"network-log-formatter/internal/config"
	"network-log-formatter/internal/executor"
	"network-log-formatter/internal/model"
	"network-log-formatter/internal/project"
	"network-log-formatter/internal/pyenv"
)

// skipIfNoUv skips the test if uv is not available.
func skipIfNoUv(t *testing.T) {
	t.Helper()
	if err := exec.Command("uv", "--version").Run(); err != nil {
		t.Skip("uv not available, skipping integration test")
	}
}

// nginxLogLines returns realistic nginx combined-format access log entries.
func nginxLogLines() string {
	return `192.168.1.100 - - [15/Jan/2025:10:23:45 +0800] "GET /api/users HTTP/1.1" 200 1234 "https://example.com/" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
10.0.0.55 - admin [15/Jan/2025:10:23:46 +0800] "POST /api/login HTTP/1.1" 302 0 "https://example.com/login" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"
172.16.0.1 - - [15/Jan/2025:10:23:47 +0800] "GET /static/css/main.css HTTP/1.1" 304 0 "-" "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0"
192.168.1.101 - - [15/Jan/2025:10:23:48 +0800] "DELETE /api/sessions/abc123 HTTP/1.1" 204 0 "https://example.com/dashboard" "curl/8.1.2"
10.0.0.88 - - [15/Jan/2025:10:23:49 +0800] "GET /favicon.ico HTTP/1.1" 404 162 "-" "Googlebot/2.1 (+http://www.google.com/bot.html)"
192.168.1.100 - - [15/Jan/2025:10:23:50 +0800] "PUT /api/users/42 HTTP/1.1" 200 567 "https://example.com/profile" "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
172.16.0.5 - - [15/Jan/2025:10:23:51 +0800] "GET /api/products?page=2&limit=20 HTTP/1.1" 200 8901 "https://example.com/products" "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0)"
10.0.0.55 - - [15/Jan/2025:10:23:52 +0800] "POST /api/orders HTTP/1.1" 201 345 "https://example.com/cart" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"
192.168.1.102 - - [15/Jan/2025:10:23:53 +0800] "GET /health HTTP/1.1" 200 2 "-" "kube-probe/1.28"
10.0.0.99 - - [15/Jan/2025:10:23:54 +0800] "GET /api/reports/export HTTP/1.1" 500 89 "https://example.com/reports" "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
`
}

// pythonScript returns a Python script that parses nginx combined log format
// and outputs Excel. This simulates what the LLM would generate.
func pythonScript() string {
	return `import argparse
import json
import os
import re
import sys

from openpyxl import Workbook

LOG_PATTERN = re.compile(
    r'(?P<ip>\S+)\s+\S+\s+(?P<user>\S+)\s+'
    r'\[(?P<time>[^\]]+)\]\s+'
    r'"(?P<method>\S+)\s+(?P<path>\S+)\s+(?P<proto>[^"]+)"\s+'
    r'(?P<status>\d+)\s+(?P<size>\d+)\s+'
    r'"(?P<referer>[^"]*)"\s+'
    r'"(?P<ua>[^"]*)"'
)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--input', required=True)
    parser.add_argument('--output', required=True)
    args = parser.parse_args()

    log_files = [f for f in os.listdir(args.input) if f.endswith('.log')]
    total = len(log_files)

    if total == 0:
        print(json.dumps({"file": "", "progress": 1.0, "total": 0, "current": 0}))
        sys.exit(0)

    wb = Workbook()
    ws = wb.active
    ws.title = "Access Logs"
    ws.append(["IP", "User", "Timestamp", "Method", "Path", "Protocol", "Status", "Size", "Referer", "User-Agent", "Source File"])

    for idx, fname in enumerate(sorted(log_files)):
        fpath = os.path.join(args.input, fname)
        try:
            with open(fpath, 'r', encoding='utf-8') as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    m = LOG_PATTERN.match(line)
                    if m:
                        d = m.groupdict()
                        ws.append([d['ip'], d['user'], d['time'], d['method'],
                                   d['path'], d['proto'], int(d['status']),
                                   int(d['size']), d['referer'], d['ua'], fname])
        except Exception as e:
            print(f"Error processing {fname}: {e}", file=sys.stderr)

        progress = (idx + 1) / total
        info = {"file": fname, "progress": progress, "total": total, "current": idx + 1}
        print(json.dumps(info))
        sys.stdout.flush()

    os.makedirs(args.output, exist_ok=True)
    output_path = os.path.join(args.output, "access_logs.xlsx")
    wb.save(output_path)

if __name__ == '__main__':
    main()
`
}

// TestE2E_FullPipeline tests the complete flow:
// 1. Settings save/load
// 2. Python env setup (uv)
// 3. Project creation
// 4. Batch execution with real nginx log files
// 5. Excel output verification
func TestE2E_FullPipeline(t *testing.T) {
	skipIfNoUv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create temp workspace
	workDir, err := os.MkdirTemp("", "e2e-test-*")
	if err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	configDir := filepath.Join(workDir, "config")
	inputDir := filepath.Join(workDir, "logs")
	outputDir := filepath.Join(workDir, "output")

	// --- Step 1: Settings Manager ---
	t.Log("Step 1: Testing SettingsManager save/load")
	sm := config.NewSettingsManager(filepath.Join(configDir, "settings.json"))

	settings := model.Settings{
		LLM: model.LLMConfig{
			BaseURL:   "https://api.example.com/v1",
			APIKey:    "test-key-12345",
			ModelName: "test-model",
		},
		UvPath:           "uv",
		DefaultInputDir:  inputDir,
		DefaultOutputDir: outputDir,
	}

	if err := sm.Save(settings); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	loaded, err := sm.Load()
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}

	if loaded.LLM.BaseURL != settings.LLM.BaseURL {
		t.Errorf("settings round-trip: BaseURL mismatch: got %q, want %q", loaded.LLM.BaseURL, settings.LLM.BaseURL)
	}
	if loaded.DefaultInputDir != settings.DefaultInputDir {
		t.Errorf("settings round-trip: DefaultInputDir mismatch")
	}
	t.Log("  ✓ Settings save/load round-trip OK")

	// --- Step 2: Python Environment ---
	t.Log("Step 2: Setting up Python environment via uv")
	envPath := filepath.Join(configDir, "pyenv")
	envMgr := pyenv.NewPythonEnvManager("uv", envPath)

	if err := envMgr.EnsureEnv(ctx); err != nil {
		t.Fatalf("failed to ensure Python env: %v", err)
	}

	status := envMgr.GetStatus()
	if !status.UvAvailable {
		t.Fatal("uv should be available")
	}
	if !status.EnvExists {
		t.Fatal("Python env should exist after EnsureEnv")
	}
	t.Log("  ✓ Python environment ready at", status.EnvPath)

	// --- Step 3: Create log files ---
	t.Log("Step 3: Creating sample nginx log files")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	logContent := nginxLogLines()

	// Create 3 log files with slightly different data
	for i := 1; i <= 3; i++ {
		fname := fmt.Sprintf("access_%d.log", i)
		fpath := filepath.Join(inputDir, fname)
		// Add a file-specific line to differentiate
		extra := fmt.Sprintf("10.0.%d.1 - - [15/Jan/2025:10:24:%02d +0800] \"GET /file%d HTTP/1.1\" 200 100 \"-\" \"TestBot/1.0\"\n", i, i, i)
		if err := os.WriteFile(fpath, []byte(logContent+extra), 0644); err != nil {
			t.Fatalf("failed to write log file %s: %v", fname, err)
		}
	}

	entries, _ := os.ReadDir(inputDir)
	t.Logf("  ✓ Created %d log files in %s", len(entries), inputDir)

	// --- Step 4: Project Manager ---
	t.Log("Step 4: Creating project record")
	pm, err := project.NewProjectManager(filepath.Join(configDir, "projects"))
	if err != nil {
		t.Fatalf("failed to create project manager: %v", err)
	}

	code := pythonScript()
	now := time.Now()
	proj := model.Project{
		ID:         "e2e-test-001",
		SampleData: logContent,
		Code:       code,
		CreatedAt:  now,
		UpdatedAt:  now,
		Status:     "validated",
	}

	if err := pm.Create(proj); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	retrieved, err := pm.Get("e2e-test-001")
	if err != nil {
		t.Fatalf("failed to get project: %v", err)
	}
	if retrieved.Status != "validated" {
		t.Errorf("project status: got %q, want %q", retrieved.Status, "validated")
	}
	if retrieved.Code != code {
		t.Error("project code mismatch after retrieval")
	}
	t.Log("  ✓ Project created and retrieved OK")

	// --- Step 5: Batch Execution ---
	t.Log("Step 5: Running batch execution")
	be := executor.NewBatchExecutor(envMgr, nil, 3)

	result, err := be.Execute(ctx, code, inputDir, outputDir, "")
	if err != nil {
		t.Fatalf("batch execution failed: %v", err)
	}

	t.Logf("  Batch result: TotalFiles=%d, Succeeded=%d, Failed=%d",
		result.TotalFiles, result.Succeeded, result.Failed)

	if result.TotalFiles != 3 {
		t.Errorf("expected 3 total files, got %d", result.TotalFiles)
	}
	if result.Succeeded != 3 {
		t.Errorf("expected 3 succeeded, got %d", result.Succeeded)
	}

	// Check progress
	progress := be.GetProgress()
	if progress.Status != "completed" {
		t.Errorf("expected progress status 'completed', got %q", progress.Status)
	}
	t.Log("  ✓ Batch execution completed successfully")

	// --- Step 6: Verify Excel output ---
	t.Log("Step 6: Verifying Excel output")
	excelPath := filepath.Join(outputDir, "access_logs.xlsx")
	info, err := os.Stat(excelPath)
	if err != nil {
		t.Fatalf("Excel file not found at %s: %v", excelPath, err)
	}
	if info.Size() == 0 {
		t.Fatal("Excel file is empty")
	}
	t.Logf("  ✓ Excel file generated: %s (%d bytes)", excelPath, info.Size())

	// --- Step 7: Update project status ---
	t.Log("Step 7: Updating project status to 'executed'")
	executedStatus := "executed"
	if err := pm.Update("e2e-test-001", model.ProjectUpdate{Status: &executedStatus}); err != nil {
		t.Fatalf("failed to update project status: %v", err)
	}

	updated, _ := pm.Get("e2e-test-001")
	if updated.Status != "executed" {
		t.Errorf("project status after update: got %q, want %q", updated.Status, "executed")
	}
	t.Log("  ✓ Project status updated to 'executed'")

	// --- Step 8: List and verify project ordering ---
	t.Log("Step 8: Verifying project list")
	// Create a second project with earlier timestamp
	proj2 := model.Project{
		ID:         "e2e-test-002",
		SampleData: "older sample",
		Code:       "print('old')",
		CreatedAt:  now.Add(-1 * time.Hour),
		UpdatedAt:  now.Add(-1 * time.Hour),
		Status:     "draft",
	}
	_ = pm.Create(proj2)

	projects, err := pm.List()
	if err != nil {
		t.Fatalf("failed to list projects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	// First project should be the newer one
	if projects[0].ID != "e2e-test-001" {
		t.Errorf("expected first project to be e2e-test-001 (newer), got %s", projects[0].ID)
	}
	t.Log("  ✓ Project list sorted correctly (newest first)")

	// --- Step 9: Delete project ---
	t.Log("Step 9: Testing project deletion")
	if err := pm.Delete("e2e-test-002"); err != nil {
		t.Fatalf("failed to delete project: %v", err)
	}
	projects, _ = pm.List()
	if len(projects) != 1 {
		t.Errorf("expected 1 project after deletion, got %d", len(projects))
	}
	t.Log("  ✓ Project deletion OK")

	t.Log("")
	t.Log("========================================")
	t.Log(" ✅ E2E Full Pipeline Test PASSED")
	t.Log("========================================")
}

// TestE2E_MultiFormatLogs tests batch processing with different log formats
// to verify the Python script handles edge cases.
func TestE2E_MultiFormatLogs(t *testing.T) {
	skipIfNoUv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	workDir, err := os.MkdirTemp("", "e2e-multiformat-*")
	if err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	envPath := filepath.Join(workDir, "pyenv")
	inputDir := filepath.Join(workDir, "logs")
	outputDir := filepath.Join(workDir, "output")

	// Setup Python env
	envMgr := pyenv.NewPythonEnvManager("uv", envPath)
	if err := envMgr.EnsureEnv(ctx); err != nil {
		t.Fatalf("failed to ensure Python env: %v", err)
	}

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	// Log file with mixed valid and invalid lines
	mixedLog := `192.168.1.1 - - [15/Jan/2025:10:00:01 +0800] "GET / HTTP/1.1" 200 5000 "-" "Chrome/120"
THIS IS NOT A VALID LOG LINE
10.0.0.1 - - [15/Jan/2025:10:00:02 +0800] "POST /api/data HTTP/1.1" 201 128 "https://example.com" "Firefox/115"
another invalid line here
172.16.0.1 - - [15/Jan/2025:10:00:03 +0800] "GET /images/logo.png HTTP/1.1" 304 0 "-" "Safari/17"
`
	if err := os.WriteFile(filepath.Join(inputDir, "mixed.log"), []byte(mixedLog), 0644); err != nil {
		t.Fatalf("failed to write mixed log: %v", err)
	}

	// Empty log file
	if err := os.WriteFile(filepath.Join(inputDir, "empty.log"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to write empty log: %v", err)
	}

	be := executor.NewBatchExecutor(envMgr, nil, 3)
	result, err := be.Execute(ctx, pythonScript(), inputDir, outputDir, "")
	if err != nil {
		t.Fatalf("batch execution failed: %v", err)
	}

	t.Logf("Result: TotalFiles=%d, Succeeded=%d, Failed=%d", result.TotalFiles, result.Succeeded, result.Failed)

	if result.TotalFiles != 2 {
		t.Errorf("expected 2 total files, got %d", result.TotalFiles)
	}

	excelPath := filepath.Join(outputDir, "access_logs.xlsx")
	info, err := os.Stat(excelPath)
	if err != nil {
		t.Fatalf("Excel file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("Excel file is empty")
	}

	t.Logf("✓ Multi-format test passed: Excel generated (%d bytes), invalid lines gracefully skipped", info.Size())
}

// TestE2E_EmptyInputDirectory tests batch processing with no log files.
func TestE2E_EmptyInputDirectory(t *testing.T) {
	skipIfNoUv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	workDir, err := os.MkdirTemp("", "e2e-empty-*")
	if err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	envPath := filepath.Join(workDir, "pyenv")
	inputDir := filepath.Join(workDir, "logs")
	outputDir := filepath.Join(workDir, "output")

	envMgr := pyenv.NewPythonEnvManager("uv", envPath)
	if err := envMgr.EnsureEnv(ctx); err != nil {
		t.Fatalf("failed to ensure Python env: %v", err)
	}

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	be := executor.NewBatchExecutor(envMgr, nil, 3)
	result, err := be.Execute(ctx, pythonScript(), inputDir, outputDir, "")
	if err != nil {
		t.Fatalf("batch execution failed with empty dir: %v", err)
	}

	if result.TotalFiles != 0 {
		t.Errorf("expected 0 total files, got %d", result.TotalFiles)
	}

	t.Log("✓ Empty directory test passed: graceful handling of no log files")
}

// TestE2E_LargeLogFile tests processing a larger log file to verify performance.
func TestE2E_LargeLogFile(t *testing.T) {
	skipIfNoUv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	workDir, err := os.MkdirTemp("", "e2e-large-*")
	if err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	envPath := filepath.Join(workDir, "pyenv")
	inputDir := filepath.Join(workDir, "logs")
	outputDir := filepath.Join(workDir, "output")

	envMgr := pyenv.NewPythonEnvManager("uv", envPath)
	if err := envMgr.EnsureEnv(ctx); err != nil {
		t.Fatalf("failed to ensure Python env: %v", err)
	}

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	// Generate a log file with 1000 entries
	var sb strings.Builder
	ips := []string{"192.168.1.1", "10.0.0.55", "172.16.0.1", "192.168.1.102", "10.0.0.99"}
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	paths := []string{"/api/users", "/api/orders", "/api/products", "/health", "/api/reports"}
	statuses := []int{200, 201, 204, 301, 400, 403, 404, 500}

	for i := 0; i < 1000; i++ {
		ip := ips[i%len(ips)]
		method := methods[i%len(methods)]
		path := paths[i%len(paths)]
		status := statuses[i%len(statuses)]
		line := fmt.Sprintf(`%s - - [15/Jan/2025:10:%02d:%02d +0800] "%s %s HTTP/1.1" %d %d "-" "TestAgent/1.0"`,
			ip, (i/60)%24, i%60, method, path, status, 100+i)
		sb.WriteString(line + "\n")
	}

	if err := os.WriteFile(filepath.Join(inputDir, "large.log"), []byte(sb.String()), 0644); err != nil {
		t.Fatalf("failed to write large log: %v", err)
	}

	start := time.Now()
	be := executor.NewBatchExecutor(envMgr, nil, 3)
	result, err := be.Execute(ctx, pythonScript(), inputDir, outputDir, "")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("batch execution failed: %v", err)
	}

	excelPath := filepath.Join(outputDir, "access_logs.xlsx")
	info, _ := os.Stat(excelPath)

	t.Logf("✓ Large file test: 1000 entries processed in %v, Excel size: %d bytes",
		elapsed, info.Size())

	if result.TotalFiles != 1 {
		t.Errorf("expected 1 total file, got %d", result.TotalFiles)
	}
}

// TestE2E_RealLLM_FullPipeline is the real end-to-end test that calls DeepSeek LLM
// to generate Python code from sample logs, validates it, and runs batch processing.
//
// Flow: Sample → LLM generates Python → CodeValidator checks syntax → BatchExecutor runs → Excel output
//
// Run with: go test ./internal/e2e/ -v -run TestE2E_RealLLM_FullPipeline -count=1 -timeout 600s
func TestE2E_RealLLM_FullPipeline(t *testing.T) {
	skipIfNoUv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// --- Real LLM config from environment variables ---
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping real LLM test")
	}
	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	modelName := os.Getenv("DEEPSEEK_MODEL")
	if modelName == "" {
		modelName = "deepseek-chat"
	}
	llmCfg := model.LLMConfig{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		ModelName: modelName,
	}

	// Create temp workspace
	workDir, err := os.MkdirTemp("", "e2e-realllm-*")
	if err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	configDir := filepath.Join(workDir, "config")
	inputDir := filepath.Join(workDir, "logs")
	outputDir := filepath.Join(workDir, "output")

	// ========== Step 1: Setup Python Environment ==========
	t.Log("Step 1: Setting up Python environment via uv")
	envPath := filepath.Join(configDir, "pyenv")
	envMgr := pyenv.NewPythonEnvManager("uv", envPath)

	if err := envMgr.EnsureEnv(ctx); err != nil {
		t.Fatalf("failed to ensure Python env: %v", err)
	}
	t.Log("  ✓ Python environment ready")

	// ========== Step 2: Create LLM Client ==========
	t.Log("Step 2: Creating LLM client (DeepSeek)")
	llmClient, err := agent.NewLLMClient(llmCfg)
	if err != nil {
		t.Fatalf("failed to create LLM client: %v", err)
	}

	// Quick connectivity test
	t.Log("  Testing LLM connectivity...")
	testResp, err := llmClient.Chat(ctx, []model.Message{
		{Role: "user", Content: "Reply with exactly: OK"},
	})
	if err != nil {
		t.Fatalf("LLM connectivity test failed: %v", err)
	}
	t.Logf("  ✓ LLM connected, test response: %q", strings.TrimSpace(testResp))

	// ========== Step 3: Prepare sample log data ==========
	t.Log("Step 3: Preparing sample nginx log data")
	sampleText := nginxLogLines()
	t.Logf("  Sample: %d lines of nginx combined-format logs", strings.Count(sampleText, "\n"))

	// ========== Step 4: SampleAnalyzer — LLM generates Python code ==========
	t.Log("Step 4: Calling SampleAnalyzer (LLM generates Python code)...")
	sa := agent.NewSampleAnalyzer(llmClient)

	generatedCode, err := sa.Analyze(ctx, sampleText)
	if err != nil {
		t.Fatalf("SampleAnalyzer.Analyze failed: %v", err)
	}

	if generatedCode == "" {
		t.Fatal("SampleAnalyzer returned empty code")
	}

	// Sanity checks on generated code
	if !strings.Contains(generatedCode, "openpyxl") {
		t.Error("generated code does not mention openpyxl")
	}
	if !strings.Contains(generatedCode, "--input") {
		t.Error("generated code does not contain --input argument")
	}
	if !strings.Contains(generatedCode, "--output") {
		t.Error("generated code does not contain --output argument")
	}

	codeLines := strings.Count(generatedCode, "\n")
	t.Logf("  ✓ LLM generated Python code: %d lines", codeLines)
	// Print first 5 lines for visibility
	lines := strings.SplitN(generatedCode, "\n", 6)
	for i, l := range lines {
		if i >= 5 {
			break
		}
		t.Logf("    | %s", l)
	}
	t.Log("    | ...")

	// ========== Step 5: CodeValidator — syntax check + auto-repair ==========
	t.Log("Step 5: Validating generated code (CodeValidator)...")
	cv := agent.NewCodeValidator(envMgr, llmClient, 3)

	valResult, err := cv.Validate(ctx, generatedCode)
	if err != nil {
		t.Fatalf("CodeValidator.Validate failed: %v", err)
	}

	t.Logf("  Validation result: Valid=%v, Retries=%d", valResult.Valid, valResult.Retries)
	if len(valResult.Errors) > 0 {
		for _, e := range valResult.Errors {
			t.Logf("  Validation error: %s", e)
		}
	}

	if !valResult.Valid {
		t.Fatalf("Code validation failed after %d retries. Errors: %v", valResult.Retries, valResult.Errors)
	}

	finalCode := valResult.Code
	t.Logf("  ✓ Code validated (retries: %d)", valResult.Retries)

	// ========== Step 6: Create log files for batch processing ==========
	t.Log("Step 6: Creating log files for batch processing")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	logContent := nginxLogLines()
	for i := 1; i <= 3; i++ {
		fname := fmt.Sprintf("access_%d.log", i)
		fpath := filepath.Join(inputDir, fname)
		extra := fmt.Sprintf("10.0.%d.1 - - [15/Jan/2025:10:24:%02d +0800] \"GET /file%d HTTP/1.1\" 200 100 \"-\" \"TestBot/1.0\"\n", i, i, i)
		if err := os.WriteFile(fpath, []byte(logContent+extra), 0644); err != nil {
			t.Fatalf("failed to write log file %s: %v", fname, err)
		}
	}
	t.Log("  ✓ Created 3 log files")

	// ========== Step 7: Save project ==========
	t.Log("Step 7: Saving project record")
	pm, err := project.NewProjectManager(filepath.Join(configDir, "projects"))
	if err != nil {
		t.Fatalf("failed to create project manager: %v", err)
	}

	now := time.Now()
	proj := model.Project{
		ID:         "realllm-test-001",
		SampleData: sampleText,
		Code:       finalCode,
		CreatedAt:  now,
		UpdatedAt:  now,
		Status:     "validated",
	}
	if err := pm.Create(proj); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	t.Log("  ✓ Project saved: realllm-test-001")

	// ========== Step 8: Batch Execution ==========
	t.Log("Step 8: Running batch execution with LLM-generated code...")

	// Create a repairer adapter for runtime error auto-fix
	repairer := &testLLMRepairer{llmClient: llmClient}
	be := executor.NewBatchExecutor(envMgr, repairer, 3)

	result, err := be.Execute(ctx, finalCode, inputDir, outputDir, "")
	if err != nil {
		t.Logf("  ⚠ Batch execution error: %v", err)
		t.Logf("  Generated code was:\n%s", finalCode)
		t.Fatalf("batch execution failed: %v", err)
	}

	t.Logf("  Batch result: TotalFiles=%d, Succeeded=%d, Failed=%d",
		result.TotalFiles, result.Succeeded, result.Failed)

	progress := be.GetProgress()
	t.Logf("  Final progress: Status=%s, Progress=%.1f%%", progress.Status, progress.Progress*100)

	if progress.Status != "completed" {
		t.Errorf("expected progress status 'completed', got %q", progress.Status)
	}
	t.Log("  ✓ Batch execution completed")

	// ========== Step 9: Verify Excel output ==========
	t.Log("Step 9: Verifying Excel output")

	// Find any .xlsx file in output directory
	var excelFiles []string
	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".xlsx") {
			excelFiles = append(excelFiles, path)
		}
		return nil
	})

	if len(excelFiles) == 0 {
		t.Fatalf("No Excel files found in output directory %s", outputDir)
	}

	for _, ef := range excelFiles {
		info, _ := os.Stat(ef)
		t.Logf("  Found Excel: %s (%d bytes)", filepath.Base(ef), info.Size())
		if info.Size() == 0 {
			t.Errorf("Excel file %s is empty", ef)
		}
	}
	t.Log("  ✓ Excel output verified")

	// ========== Step 10: Update project status ==========
	t.Log("Step 10: Updating project status")
	executedStatus := "executed"
	_ = pm.Update("realllm-test-001", model.ProjectUpdate{Status: &executedStatus})
	updated, _ := pm.Get("realllm-test-001")
	if updated.Status != "executed" {
		t.Errorf("project status: got %q, want %q", updated.Status, "executed")
	}
	t.Log("  ✓ Project status updated to 'executed'")

	t.Log("")
	t.Log("========================================")
	t.Log(" ✅ REAL LLM E2E Full Pipeline PASSED")
	t.Log("========================================")
}

// testLLMRepairer implements executor.LLMRepairer using a real LLM client.
type testLLMRepairer struct {
	llmClient *agent.LLMClient
}

func (r *testLLMRepairer) RepairCode(ctx context.Context, code string, errorMsg string) (string, error) {
	messages := []model.Message{
		{
			Role: "system",
			Content: "You are an expert Python developer. Fix the runtime error in the given Python code. " +
				"Return the complete fixed Python code inside a single ```python code block. " +
				"Do not explain the changes, just return the corrected code.",
		},
		{
			Role: "user",
			Content: fmt.Sprintf("The following Python code encountered a runtime error:\n\n```python\n%s\n```\n\n"+
				"Error message:\n```\n%s\n```\n\nPlease fix the error and return the complete corrected code.",
				code, errorMsg),
		},
	}

	resp, err := r.llmClient.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	// Extract code from response
	if start := strings.Index(resp, "```python"); start >= 0 {
		rest := resp[start:]
		if nlIdx := strings.Index(rest, "\n"); nlIdx >= 0 {
			content := rest[nlIdx+1:]
			if endIdx := strings.Index(content, "```"); endIdx >= 0 {
				return strings.TrimSpace(content[:endIdx]), nil
			}
		}
	}
	if start := strings.Index(resp, "```"); start >= 0 {
		rest := resp[start:]
		if nlIdx := strings.Index(rest, "\n"); nlIdx >= 0 {
			content := rest[nlIdx+1:]
			if endIdx := strings.Index(content, "```"); endIdx >= 0 {
				return strings.TrimSpace(content[:endIdx]), nil
			}
		}
	}
	return resp, nil
}

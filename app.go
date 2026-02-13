package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"network-log-formatter/internal/agent"
	"network-log-formatter/internal/config"
	"network-log-formatter/internal/executor"
	"network-log-formatter/internal/model"
	"network-log-formatter/internal/project"
	"network-log-formatter/internal/pyenv"
)

// App is the main controller bridging the Wails frontend and Go backend.
type App struct {
	ctx             context.Context
	configDir       string
	sampleAnalyzer  *agent.SampleAnalyzer
	codeValidator   *agent.CodeValidator
	batchExecutor   *executor.BatchExecutor
	envManager      *pyenv.PythonEnvManager
	projectManager  *project.ProjectManager
	settingsManager *config.SettingsManager
	llmClient       *agent.LLMClient
	mu              sync.Mutex // protects pyenvReady and pyenvError
	pyenvReady      bool
	pyenvError      string
}

// NewApp creates a new App with SettingsManager and ProjectManager initialized.
// LLM-dependent components are initialized lazily after settings are loaded.
func NewApp(configDir string) *App {
	settingsMgr := config.NewSettingsManager(filepath.Join(configDir, "settings.json"))
	projectsDir := filepath.Join(configDir, "projects")
	projectMgr, err := project.NewProjectManager(projectsDir)
	if err != nil {
		// Log but don't fail — projectManager will be nil and methods will return errors
		fmt.Printf("warning: failed to initialize project manager: %v\n", err)
	}

	return &App{
		configDir:       configDir,
		settingsManager: settingsMgr,
		projectManager:  projectMgr,
	}
}

// startup is called by Wails when the application starts.
// It loads settings and initializes LLM-dependent components.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	settings, err := a.settingsManager.Load()
	if err != nil {
		fmt.Printf("warning: failed to load settings: %v\n", err)
		// Still initialize env manager with defaults so the app is usable
		envPath := filepath.Join(a.configDir, "pyenv")
		a.envManager = pyenv.NewPythonEnvManager("uv", envPath)
		return
	}

	// Initialize Python environment manager
	uvPath := settings.UvPath
	if uvPath == "" {
		uvPath = "uv"
	}
	envPath := filepath.Join(a.configDir, "pyenv")
	a.envManager = pyenv.NewPythonEnvManager(uvPath, envPath)

	// Auto-initialize Python environment in background
	go func() {
		if err := a.envManager.EnsureEnv(a.ctx); err != nil {
			a.mu.Lock()
			a.pyenvError = err.Error()
			a.mu.Unlock()
			fmt.Printf("warning: auto python env init failed: %v\n", err)
		} else {
			a.mu.Lock()
			a.pyenvReady = true
			a.mu.Unlock()
		}
	}()

	// Initialize LLM components if configured
	if settings.LLM.BaseURL != "" && settings.LLM.APIKey != "" && settings.LLM.ModelName != "" {
		_ = a.initLLMComponents(settings.LLM)
	}
}

// initLLMComponents initializes or reinitializes the LLM client and all
// components that depend on it (SampleAnalyzer, CodeValidator, BatchExecutor).
func (a *App) initLLMComponents(cfg model.LLMConfig) error {
	llmClient, err := agent.NewLLMClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}
	a.llmClient = llmClient
	a.sampleAnalyzer = agent.NewSampleAnalyzer(llmClient)
	a.codeValidator = agent.NewCodeValidator(a.envManager, llmClient, 3)
	a.batchExecutor = executor.NewBatchExecutor(
		a.envManager,
		&llmRepairerAdapter{llmClient: llmClient},
		3,
	)
	return nil
}

// AnalyzeSample analyzes sample log text: calls SampleAnalyzer → CodeValidator → ProjectManager.
func (a *App) AnalyzeSample(projectName string, sampleText string) (*model.GenerateResult, error) {
	if a.sampleAnalyzer == nil {
		return nil, fmt.Errorf("LLM is not configured. Please configure LLM settings first")
	}

	if strings.TrimSpace(projectName) == "" {
		return nil, fmt.Errorf("请输入项目名称")
	}

	// 1. Analyze sample to generate Python code
	analyzeCtx, analyzeCancel := context.WithTimeout(a.ctx, 2*time.Minute)
	defer analyzeCancel()
	code, err := a.sampleAnalyzer.Analyze(analyzeCtx, sampleText)
	if err != nil {
		return nil, fmt.Errorf("sample analysis failed: %w", err)
	}

	// 2. Validate the generated code
	var validationResult *agent.ValidationResult
	if a.codeValidator != nil {
		validationResult, err = a.codeValidator.Validate(a.ctx, code)
		if err != nil {
			return nil, fmt.Errorf("code validation failed: %w", err)
		}
		code = validationResult.Code
	}

	// 3. Determine project status
	status := "draft"
	valid := false
	var errors []string
	if validationResult != nil {
		valid = validationResult.Valid
		errors = validationResult.Errors
		if valid {
			status = "validated"
		}
	}

	// 4. Create project
	projectID := uuid.New().String()
	now := time.Now()
	p := model.Project{
		ID:         projectID,
		Name:       strings.TrimSpace(projectName),
		SampleData: sampleText,
		Code:       code,
		CreatedAt:  now,
		UpdatedAt:  now,
		Status:     status,
	}

	if a.projectManager != nil {
		if err := a.projectManager.Create(p); err != nil {
			return nil, fmt.Errorf("failed to save project: %w", err)
		}
	}

	return &model.GenerateResult{
		ProjectID: projectID,
		Code:      code,
		Valid:     valid,
		Errors:    errors,
	}, nil
}

// RunBatch starts batch processing in a background goroutine so it doesn't block the UI.
func (a *App) RunBatch(projectID string, inputDir string, outputDir string) error {
	if a.batchExecutor == nil {
		return fmt.Errorf("LLM is not configured. Please configure LLM settings first")
	}
	if a.projectManager == nil {
		return fmt.Errorf("project manager is not initialized")
	}

	a.mu.Lock()
	envReady := a.pyenvReady
	a.mu.Unlock()
	if !envReady {
		return fmt.Errorf("Python 环境尚未就绪，请等待初始化完成")
	}

	p, err := a.projectManager.Get(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if strings.TrimSpace(p.Code) == "" {
		return fmt.Errorf("项目代码为空，无法执行")
	}

	go func() {
		_, execErr := a.batchExecutor.Execute(a.ctx, p.Code, inputDir, outputDir)

		// Update project status based on result
		status := "executed"
		if execErr != nil {
			status = "failed"
		}
		statusStr := status
		_ = a.projectManager.Update(projectID, model.ProjectUpdate{Status: &statusStr})
	}()

	return nil
}

// GetBatchProgress returns the current batch processing progress.
func (a *App) GetBatchProgress() (*model.BatchProgress, error) {
	if a.batchExecutor == nil {
		return &model.BatchProgress{Status: "idle"}, nil
	}
	return a.batchExecutor.GetProgress(), nil
}

// ListProjects returns all projects sorted by creation time descending.
func (a *App) ListProjects() ([]model.Project, error) {
	if a.projectManager == nil {
		return nil, fmt.Errorf("project manager is not initialized")
	}
	return a.projectManager.List()
}

// GetProject returns a single project by ID.
func (a *App) GetProject(id string) (*model.Project, error) {
	if a.projectManager == nil {
		return nil, fmt.Errorf("project manager is not initialized")
	}
	return a.projectManager.Get(id)
}

// UpdateProjectCode updates the Python code for a project.
func (a *App) UpdateProjectCode(id string, code string) error {
	if a.projectManager == nil {
		return fmt.Errorf("project manager is not initialized")
	}
	return a.projectManager.Update(id, model.ProjectUpdate{Code: &code})
}

// DeleteProject removes a project by ID.
func (a *App) DeleteProject(id string) error {
	if a.projectManager == nil {
		return fmt.Errorf("project manager is not initialized")
	}
	return a.projectManager.Delete(id)
}

// RerunProject starts batch processing using an existing project's code.
func (a *App) RerunProject(id string, inputDir string, outputDir string) error {
	return a.RunBatch(id, inputDir, outputDir)
}

// SelectDirectory opens a native directory picker dialog and returns the selected path.
func (a *App) SelectDirectory(title string) (string, error) {
	if title == "" {
		title = "选择目录"
	}
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: title,
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

// GetSettings returns the current application settings.
func (a *App) GetSettings() (*model.Settings, error) {
	return a.settingsManager.Load()
}

// SaveSettings saves settings and reinitializes LLM-dependent components.
func (a *App) SaveSettings(settings model.Settings) error {
	if err := a.settingsManager.Save(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Reinitialize Python environment manager with new uv path
	uvPath := settings.UvPath
	if uvPath == "" {
		uvPath = "uv"
	}
	envPath := filepath.Join(a.configDir, "pyenv")
	a.envManager = pyenv.NewPythonEnvManager(uvPath, envPath)

	// Reinitialize LLM components if configured
	if settings.LLM.BaseURL != "" && settings.LLM.APIKey != "" && settings.LLM.ModelName != "" {
		if err := a.initLLMComponents(settings.LLM); err != nil {
			return fmt.Errorf("failed to reinitialize LLM components: %w", err)
		}
	}

	return nil
}

// EnsurePythonEnv ensures the Python virtual environment is set up.
func (a *App) EnsurePythonEnv() error {
	if a.envManager == nil {
		return fmt.Errorf("Python environment manager is not initialized. Please check settings")
	}
	return a.envManager.EnsureEnv(a.ctx)
}

// GetEnvStatus returns the current Python environment status.
func (a *App) GetEnvStatus() (*pyenv.EnvStatus, error) {
	if a.envManager == nil {
		return nil, fmt.Errorf("Python environment manager is not initialized")
	}
	return a.envManager.GetStatus(), nil
}

// GetPythonEnvReady returns whether the auto-init has completed and any error message.
func (a *App) GetPythonEnvReady() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	return map[string]interface{}{
		"ready": a.pyenvReady,
		"error": a.pyenvError,
	}
}

// IsLLMConfigured returns true if LLM settings are filled and the client is initialized.
func (a *App) IsLLMConfigured() bool {
	return a.llmClient != nil
}

// GetShowWizard returns whether the startup wizard should be shown.
func (a *App) GetShowWizard() bool {
	settings, err := a.settingsManager.Load()
	if err != nil {
		return true
	}
	if settings.ShowWizard == nil {
		return true
	}
	return *settings.ShowWizard
}

// SetShowWizard updates the show_wizard setting.
func (a *App) SetShowWizard(show bool) error {
	settings, err := a.settingsManager.Load()
	if err != nil {
		return err
	}
	settings.ShowWizard = &show
	return a.settingsManager.Save(*settings)
}

// TestLLM tests the LLM connection by sending a simple message and checking for a response.
func (a *App) TestLLM() error {
	settings, err := a.settingsManager.Load()
	if err != nil {
		return fmt.Errorf("无法加载设置: %w", err)
	}
	if settings.LLM.BaseURL == "" || settings.LLM.APIKey == "" || settings.LLM.ModelName == "" {
		return fmt.Errorf("LLM 配置不完整，请填写 Base URL、API Key 和 Model Name")
	}

	// Create a temporary client to test
	testClient, err := agent.NewLLMClient(settings.LLM)
	if err != nil {
		return fmt.Errorf("创建 LLM 客户端失败: %w", err)
	}

	// Use a timeout context for the test request
	testCtx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	_, err = testClient.Chat(testCtx, []model.Message{
		{Role: "user", Content: "请回复 OK"},
	})
	if err != nil {
		return fmt.Errorf("LLM 连接测试失败: %w", err)
	}

	// Test passed — initialize components with this config
	if initErr := a.initLLMComponents(settings.LLM); initErr != nil {
		return fmt.Errorf("初始化 LLM 组件失败: %w", initErr)
	}

	return nil
}

// llmRepairerAdapter adapts LLMClient to the executor.LLMRepairer interface.
type llmRepairerAdapter struct {
	llmClient *agent.LLMClient
}

func (a *llmRepairerAdapter) RepairCode(ctx context.Context, code string, errorMsg string) (string, error) {
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

	resp, err := a.llmClient.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	return extractRepairCode(resp), nil
}

// extractRepairCode extracts Python code from an LLM repair response.
func extractRepairCode(response string) string {
	// Try ```python block
	if start := strings.Index(response, "```python"); start >= 0 {
		rest := response[start:]
		if nlIdx := strings.Index(rest, "\n"); nlIdx >= 0 {
			content := rest[nlIdx+1:]
			if endIdx := strings.Index(content, "```"); endIdx >= 0 {
				return strings.TrimSpace(content[:endIdx])
			}
		}
	}
	// Try any ``` block
	if start := strings.Index(response, "```"); start >= 0 {
		rest := response[start:]
		if nlIdx := strings.Index(rest, "\n"); nlIdx >= 0 {
			content := rest[nlIdx+1:]
			if endIdx := strings.Index(content, "```"); endIdx >= 0 {
				return strings.TrimSpace(content[:endIdx])
			}
		}
	}
	return response
}

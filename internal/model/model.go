// Package model defines shared data types used across the application.
package model

import "time"

// LLMConfig holds configuration for the LLM API connection.
type LLMConfig struct {
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
}

// Settings holds global application settings.
type Settings struct {
	LLM              LLMConfig `json:"llm"`
	UvPath           string    `json:"uv_path"`
	DefaultInputDir  string    `json:"default_input_dir"`
	DefaultOutputDir string    `json:"default_output_dir"`
	ShowWizard       *bool     `json:"show_wizard,omitempty"`
}


// Project represents a single code generation project record.
type Project struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	SampleData string    `json:"sample_data"`
	Code       string    `json:"code"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Status     string    `json:"status"` // "draft", "validated", "executed", "failed"
}

// ProjectUpdate holds optional fields for partial project updates.
type ProjectUpdate struct {
	Name   *string `json:"name,omitempty"`
	Code   *string `json:"code,omitempty"`
	Status *string `json:"status,omitempty"`
}

// GenerateResult holds the result of a code generation operation.
type GenerateResult struct {
	ProjectID string   `json:"project_id"`
	Code      string   `json:"code"`
	Valid     bool     `json:"valid"`
	Errors    []string `json:"errors,omitempty"`
}

// BatchResult holds the summary of a batch processing run.
type BatchResult struct {
	TotalFiles int      `json:"total_files"`
	Succeeded  int      `json:"succeeded"`
	Failed     int      `json:"failed"`
	OutputPath string   `json:"output_path"`
	Errors     []string `json:"errors,omitempty"`
}

// BatchProgress holds the current state of a batch processing operation.
type BatchProgress struct {
	Status      string  `json:"status"` // "running", "completed", "failed", "fixing"
	CurrentFile string  `json:"current_file"`
	Progress    float64 `json:"progress"`
	TotalFiles  int     `json:"total_files"`
	Processed   int     `json:"processed"`
	Failed      int     `json:"failed"`
	Message     string  `json:"message"`
}

// Message represents a single message in an LLM conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ProgressInfo represents progress output from the Python processing script (stdout JSON).
type ProgressInfo struct {
	File     string  `json:"file"`
	Progress float64 `json:"progress"`
	Total    int     `json:"total"`
	Current  int     `json:"current"`
}

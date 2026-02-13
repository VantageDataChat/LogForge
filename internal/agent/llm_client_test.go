package agent

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"

	"network-log-formatter/internal/model"
)

func TestNewLLMClient_EmptyBaseURL(t *testing.T) {
	_, err := NewLLMClient(model.LLMConfig{
		BaseURL:   "",
		APIKey:    "sk-test",
		ModelName: "test-model",
	})
	if err == nil {
		t.Fatal("expected error for empty BaseURL")
	}
}

func TestNewLLMClient_EmptyAPIKey(t *testing.T) {
	_, err := NewLLMClient(model.LLMConfig{
		BaseURL:   "https://api.example.com",
		APIKey:    "",
		ModelName: "test-model",
	})
	if err == nil {
		t.Fatal("expected error for empty APIKey")
	}
}

func TestNewLLMClient_EmptyModelName(t *testing.T) {
	_, err := NewLLMClient(model.LLMConfig{
		BaseURL:   "https://api.example.com",
		APIKey:    "sk-test",
		ModelName: "",
	})
	if err == nil {
		t.Fatal("expected error for empty ModelName")
	}
}

func TestNewLLMClient_ValidConfig(t *testing.T) {
	client, err := NewLLMClient(model.LLMConfig{
		BaseURL:   "https://api.example.com",
		APIKey:    "sk-test",
		ModelName: "test-model",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.chatModel == nil {
		t.Fatal("expected non-nil chatModel")
	}
}

func TestChat_EmptyMessages(t *testing.T) {
	client, err := NewLLMClient(model.LLMConfig{
		BaseURL:   "https://api.example.com",
		APIKey:    "sk-test",
		ModelName: "test-model",
	})
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	_, err = client.Chat(context.Background(), []model.Message{})
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
}

func TestChat_NilMessages(t *testing.T) {
	client, err := NewLLMClient(model.LLMConfig{
		BaseURL:   "https://api.example.com",
		APIKey:    "sk-test",
		ModelName: "test-model",
	})
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	_, err = client.Chat(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil messages")
	}
}

func TestConvertRole(t *testing.T) {
	tests := []struct {
		input    string
		expected schema.RoleType
	}{
		{"system", schema.System},
		{"user", schema.User},
		{"assistant", schema.Assistant},
		{"unknown", schema.User},
		{"", schema.User},
	}

	for _, tt := range tests {
		got := convertRole(tt.input)
		if got != tt.expected {
			t.Errorf("convertRole(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

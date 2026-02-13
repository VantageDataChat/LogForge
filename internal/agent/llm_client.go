// Package agent provides LLM integration components.
package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	appmodel "network-log-formatter/internal/model"
)

// LLMClient wraps the Eino framework for LLM communication.
// Supports configurable BaseURL, APIKey, and ModelName.
type LLMClient struct {
	chatModel model.ChatModel
}

// NewLLMClient creates a new LLMClient with the given configuration.
// It validates that BaseURL, APIKey, and ModelName are all non-empty.
func NewLLMClient(cfg appmodel.LLMConfig) (*LLMClient, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("LLM BaseURL must not be empty")
	}
	if cfg.APIKey == "" {
		return nil, errors.New("LLM APIKey must not be empty")
	}
	if cfg.ModelName == "" {
		return nil, errors.New("LLM ModelName must not be empty")
	}

	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Model:   cfg.ModelName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Eino ChatModel: %w", err)
	}

	return &LLMClient{chatModel: chatModel}, nil
}

// Chat sends messages to the LLM and returns the response content.
func (c *LLMClient) Chat(ctx context.Context, messages []appmodel.Message) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("messages must not be empty")
	}

	schemaMessages := make([]*schema.Message, len(messages))
	for i, msg := range messages {
		schemaMessages[i] = &schema.Message{
			Role:    convertRole(msg.Role),
			Content: msg.Content,
		}
	}

	resp, err := c.chatModel.Generate(ctx, schemaMessages)
	if err != nil {
		return "", fmt.Errorf("LLM generate failed: %w", err)
	}

	return resp.Content, nil
}

// convertRole maps a string role to the Eino schema RoleType.
func convertRole(role string) schema.RoleType {
	switch role {
	case "system":
		return schema.System
	case "assistant":
		return schema.Assistant
	case "user":
		return schema.User
	default:
		return schema.User
	}
}

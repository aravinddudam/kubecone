package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

const defaultOpenAIModel = "gpt-4o-mini"
const defaultOpenAIBase = "https://api.openai.com/v1"

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type OpenAI struct {
	APIKey  string
	Model   string
	BaseURL string
	Client  HTTPDoer
}

func (OpenAI) Name() string { return "openai" }

func NewOpenAIFromEnv() OpenAI {
	return OpenAI{
		APIKey:  strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		Model:   envOr("OPENAI_MODEL", defaultOpenAIModel),
		BaseURL: strings.TrimRight(envOr("OPENAI_BASE_URL", defaultOpenAIBase), "/"),
	}
}

func (o OpenAI) Diagnose(ctx context.Context, report *evidence.Report) (*evidence.AINote, error) {
	if strings.TrimSpace(o.APIKey) == "" {
		return &evidence.AINote{
			Provider: o.Name(),
			Skipped:  true,
			Reason:   "OPENAI_API_KEY is not set",
		}, nil
	}
	model := o.Model
	if model == "" {
		model = defaultOpenAIModel
	}
	text, err := o.complete(ctx, systemPrompt(), BuildPrompt(report), model)
	if err != nil {
		return &evidence.AINote{
			Provider: o.Name(),
			Skipped:  true,
			Reason:   err.Error(),
		}, nil
	}
	return &evidence.AINote{
		Provider: o.Name(),
		Text:     text,
	}, nil
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (o OpenAI) complete(ctx context.Context, system, user, model string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.2,
		MaxTokens:   700,
	})
	if err != nil {
		return "", err
	}

	base := o.BaseURL
	if base == "" {
		base = defaultOpenAIBase
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+o.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := o.Client
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("openai read: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("openai decode: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("openai: %s", parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai HTTP %d: %s", resp.StatusCode, clip(string(raw), 240))
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("openai returned an empty completion")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

package ai

import (
	"context"
	"errors"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type Anthropic struct {
	APIKey string
}

func (Anthropic) Name() string { return "anthropic" }

func (a Anthropic) Diagnose(context.Context, *evidence.Report) (*evidence.AINote, error) {
	if a.APIKey == "" {
		return &evidence.AINote{
			Provider: a.Name(),
			Skipped:  true,
			Reason:   "ANTHROPIC_API_KEY is not set",
		}, nil
	}
	return nil, errors.New("anthropic provider is stubbed until the evidence pipeline is proven")
}

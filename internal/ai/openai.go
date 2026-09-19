package ai

import (
	"context"
	"errors"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type OpenAI struct {
	APIKey string
}

func (OpenAI) Name() string { return "openai" }

func (o OpenAI) Diagnose(context.Context, *evidence.Report) (*evidence.AINote, error) {
	if o.APIKey == "" {
		return &evidence.AINote{
			Provider: o.Name(),
			Skipped:  true,
			Reason:   "OPENAI_API_KEY is not set",
		}, nil
	}
	return nil, errors.New("openai provider is stubbed until the evidence pipeline is proven")
}

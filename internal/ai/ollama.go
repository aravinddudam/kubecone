package ai

import (
	"context"
	"errors"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type Ollama struct {
	Host string
}

func (Ollama) Name() string { return "ollama" }

func (o Ollama) Diagnose(context.Context, *evidence.Report) (*evidence.AINote, error) {
	if o.Host == "" {
		return &evidence.AINote{
			Provider: o.Name(),
			Skipped:  true,
			Reason:   "OLLAMA_HOST is not set",
		}, nil
	}
	return nil, errors.New("ollama provider is stubbed until the evidence pipeline is proven")
}

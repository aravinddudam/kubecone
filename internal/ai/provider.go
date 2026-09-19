package ai

import (
	"context"
	"os"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

// Provider is the optional last step of the pipeline. A model is called only
// when --ai is set. Ranked findings are sent, not a blank "what's wrong?" prompt.
type Provider interface {
	Name() string
	Diagnose(ctx context.Context, report *evidence.Report) (*evidence.AINote, error)
}

func Lookup(name string) Provider {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "openai", "":
		return NewOpenAIFromEnv()
	case "anthropic":
		return Anthropic{APIKey: os.Getenv("ANTHROPIC_API_KEY")}
	case "ollama":
		return Ollama{Host: strings.TrimSpace(os.Getenv("OLLAMA_HOST"))}
	default:
		return Disabled{}
	}
}

type Disabled struct{}

func (Disabled) Name() string { return "none" }

func (Disabled) Diagnose(context.Context, *evidence.Report) (*evidence.AINote, error) {
	return &evidence.AINote{
		Provider: "none",
		Skipped:  true,
		Reason:   "deterministic diagnosis only; pass --ai after configuring a provider",
	}, nil
}

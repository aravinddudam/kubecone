package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/aravinddudam/kubecone/internal/ai"
	"github.com/aravinddudam/kubecone/internal/analyzer"
	"github.com/aravinddudam/kubecone/internal/collector"
	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

type Options struct {
	Kubeconfig string
	Context    string
	Namespace  string
	LogTail    int64
	Timeout    time.Duration
	AI         bool
	AIProvider string
}

type Engine struct {
	collect  func(ctx context.Context, target kubernetes.Target) (*evidence.Snapshot, error)
	analyze  func(snap evidence.Snapshot) []evidence.Finding
	ai       ai.Provider
}

func New(opts Options) (*Engine, error) {
	client, err := kubernetes.New(kubernetes.Options{
		Kubeconfig: opts.Kubeconfig,
		Context:    opts.Context,
		Namespace:  opts.Namespace,
	})
	if err != nil {
		return nil, err
	}
	col := collector.New(client, collector.Options{LogTail: opts.LogTail})
	eng := &Engine{
		collect: col.Collect,
		analyze: func(snap evidence.Snapshot) []evidence.Finding {
			return analyzer.Run(snap, nil)
		},
		ai: ai.Lookup(opts.AIProvider),
	}
	if !opts.AI {
		eng.ai = ai.Disabled{}
	}
	return eng, nil
}

func NewWithCollector(collect func(ctx context.Context, target kubernetes.Target) (*evidence.Snapshot, error), analyzers []analyzer.Analyzer) *Engine {
	return &Engine{
		collect: collect,
		analyze: func(snap evidence.Snapshot) []evidence.Finding {
			return analyzer.Run(snap, analyzers)
		},
		ai: ai.Disabled{},
	}
}

func (e *Engine) Investigate(ctx context.Context, target kubernetes.Target) (*evidence.Report, error) {
	snap, err := e.collect(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("collect evidence: %w", err)
	}
	findings := e.analyze(*snap)
	if len(findings) == 0 {
		findings = []evidence.Finding{{
			Code:           evidence.CodeHealthy,
			Severity:       evidence.SevInfo,
			Title:          "No deterministic failure found",
			Summary:        "KubeCone collected cluster evidence and did not match OOM, crash loop, image pull, scheduling, or probe rules.",
			Resource:       target.String(),
			Recommendation: "If the workload is still unhealthy, inspect recent events and consider enabling AI enrichment once a provider is configured.",
			Confidence:     0.4,
		}}
	}

	report := &evidence.Report{
		Target:      snap.Target,
		Cluster:     snap.Cluster,
		CollectedAt: snap.CollectedAt,
		Primary:     analyzer.Primary(findings),
		Findings:    findings,
		Graph:       evidence.BuildGraph(*snap),
		Evidence:    *snap,
	}
	if e.ai != nil {
		note, err := e.ai.Diagnose(ctx, report)
		if err != nil {
			report.AI = &evidence.AINote{Provider: e.ai.Name(), Skipped: true, Reason: err.Error()}
		} else {
			report.AI = note
		}
	}
	return report, nil
}

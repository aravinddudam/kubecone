# Architecture

KubeCone is a Go CLI. It investigates a Kubernetes resource and prints a diagnosis.

The v0.2 pipeline is deterministic. An LLM is an optional last step, not the product.

```
kubecone investigate deployment/payment-api
        │
        ▼
  client-go clientset
        │
        ▼
     Collector
   pods, events, logs,
   workload, services
        │
        ▼
   Evidence Snapshot
        │
        ├──► Evidence Graph
        │
        ▼
    Rule Engine
  image pull, scheduling,
  oom, probe, crash loop
        │
        ▼
     Diagnosis
   JSON + terminal
        │
        ▼
   AI provider (opt-in)
```

## Packages

| Package | Role |
| --- | --- |
| `internal/kubernetes` | client-go access layer and `kind/name` target parsing |
| `internal/collector` | one-shot collection into an `evidence.Snapshot` |
| `internal/evidence` | snapshot, graph, finding, report types |
| `internal/analyzer` | deterministic rules and ranking |
| `internal/engine` | investigate orchestration and namespace scan |
| `internal/reporter` | text and JSON output |
| `internal/cli` | Cobra commands: investigate, scan, pods, events, nodes, namespaces, current-context, explain, version |
| `internal/ai` | optional OpenAI enrichment after ranked findings; anthropic/ollama still stubs |

## What this borrowed, and what it did not copy

KubeCone is original code. Nearby clones were used as study material:

- **K9s**: Cobra CLI layout, kubeconfig flags, `internal/` packaging
- **Robusta**: collect context, enrich, then act; OOM and crash-loop as first-class signals
- **KRR**: recommendations should say what to change (memory bump after OOM)
- **client-go**: the actual cluster API
- **kind**: disposable cluster for the incident lab
- **Metrics Server / OpenTelemetry / controller-runtime**: documented future seams, not dependencies yet

controller-runtime informers, Prometheus, and OpenTelemetry are intentionally absent until there is a concrete need.

## Ranking

More specific failures win over generic ones. An OOMKilled container that later sits in CrashLoopBackOff is reported as `OOM_KILLED`, not `CRASH_LOOP`.

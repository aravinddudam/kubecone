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

## Ranking

More specific failures win over generic ones. An OOMKilled container that later sits in CrashLoopBackOff is reported as `OOM_KILLED`, not `CRASH_LOOP`.

<<<<<<< HEAD
# kubecone
kubecone v0.1
=======
# KubeCone

Deterministic Kubernetes investigation CLI.

```bash
kubecone investigate deployment/payment-api
```

KubeCone collects cluster evidence, builds an evidence graph, identifies the failure with rules, then prints a terminal report or JSON. AI is opt-in and comes after the diagnosis, not before it.

## Status

v0.1 diagnoses these failures without a model:

| Code | Meaning |
| --- | --- |
| `OOM_KILLED` | container exceeded its memory limit |
| `CRASH_LOOP` | container is restarting |
| `IMAGE_PULL` | image cannot be pulled |
| `FAILED_SCHEDULING` | pod never bound to a node |
| `PROBE_FAILED` | readiness/liveness probe failing |

## Install

Requires Go 1.23+ and a kubeconfig that `kubectl` already uses.

```bash
cd kubecone
go build -o bin/kubecone ./cmd/kubecone
```

## Usage

```bash
kubecone investigate deployment/payment-api
kubecone investigate deploy/payment-api -n payments -o json
kubecone investigate pod/payment-api-5d9c7b -o json
kubecone version
```

Flags: `--kubeconfig`, `--context`, `--namespace` / `-n`, `--output` / `-o`, `--timeout`, `--log-tail`, `--ai`, `--ai-provider`.

`--ai` does nothing useful yet. Providers are stubs so the pipeline stays honest: collect, graph, rules, report.

## Pipeline

```
Target
  → Kubernetes evidence
  → Evidence graph
  → Deterministic analyzers
  → Structured JSON / terminal report
  → Optional AI (not in v0.1)
```

See [docs/architecture.md](docs/architecture.md).

## Test lab

Broken workloads live in `tests/incidents/`. They are meant to run on [kind](https://kind.sigs.k8s.io/):

```bash
kind create cluster --name kubecone
kubectl apply -f tests/incidents/namespace.yaml
kubectl apply -f tests/incidents/crash-loop/manifest.yaml
./bin/kubecone investigate deployment/checkout-worker -n kubecone-lab
```

Unit tests do not need a cluster:

```bash
go test ./...
```

## Dependencies

v0.1 keeps the module small:

- `k8s.io/client-go`
- `k8s.io/apimachinery`
- `k8s.io/api`
- `github.com/spf13/cobra`

Prometheus, OpenTelemetry, and controller-runtime are later.

## Layout

```
cmd/kubecone/           CLI entrypoint
internal/kubernetes/    client-go access layer
internal/collector/     pods, events, logs, metrics seam
internal/evidence/      snapshot, graph, findings
internal/analyzer/      oom, crashloop, imagepull, scheduling, probe
internal/ai/            provider stubs
internal/reporter/      terminal + JSON
tests/incidents/        broken-app fixtures for kind
```
>>>>>>> b68d009 (Initial commit)

# KubeCone

[![CI](https://github.com/aravinddudam/kubecone/actions/workflows/ci.yml/badge.svg)](https://github.com/aravinddudam/kubecone/actions/workflows/ci.yml)
[![Go 1.23+](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Deterministic Kubernetes investigation CLI. It collects cluster evidence, builds an evidence graph, ranks failures with rules, and prints a terminal report or JSON.

AI is optional and runs **after** the diagnosis. v0.1 does not call a model unless you pass `--ai`, and even then providers are stubs.

```bash
kubecone scan -n banking-dev --unhealthy
kubecone investigate deployment/customer-service -n banking-dev
kubecone explain IMAGE_PULL
```

## Table of contents

- [What it does](#what-it-does)
- [Prerequisites](#prerequisites)
- [Install](#install)
- [Cluster access](#cluster-access)
- [Quick start](#quick-start)
- [Commands](#commands)
- [Flags](#flags)
- [Output and exit codes](#output-and-exit-codes)
- [Diagnosis codes](#diagnosis-codes)
- [Develop and test](#develop-and-test)
- [Repository layout](#repository-layout)
- [License](#license)

## What it does

KubeCone talks to the same API server `kubectl` uses. Point it at a Deployment, Pod, or Service and it:

1. Collects pods, events, logs, and related objects
2. Builds an evidence graph
3. Runs deterministic analyzers (image pull, scheduling, OOM, probes, crash loop)
4. Prints a ranked primary diagnosis plus related findings

It does **not** mutate the cluster. It is read-only.

## Prerequisites

You need these before you can build or run KubeCone:

| Requirement | Version / notes |
| --- | --- |
| [Go](https://go.dev/dl/) | **1.23 or newer** (`go version`) |
| [Git](https://git-scm.com/) | to clone this repository |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | same kubeconfig KubeCone will use |
| Kubernetes cluster | any cluster `kubectl get nodes` can reach (k3s, kind, EKS, …) |
| RBAC | permission to **get/list** pods, deployments, events, logs, and services in the target namespace |

Optional:

| Tool | When you need it |
| --- | --- |
| [kind](https://kind.sigs.k8s.io/) | local broken-workload lab in `tests/incidents/` |
| `make` | convenience targets (`make build`, `make test`) |
| Docker | only if you use kind |

Check the machine:

```bash
go version          # go1.23 or later
git --version
kubectl version --client
kubectl get nodes   # must succeed against your cluster
```

If `go` is missing on Amazon Linux / EC2, install the official toolchain (not the distro `golang` package):

```bash
# x86_64
curl -L -o /tmp/go.tgz https://go.dev/dl/go1.23.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf /tmp/go.tgz
echo 'export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
go version
```

Use `linux-arm64` instead of `linux-amd64` when `uname -m` is `aarch64`. Newer Go 1.23+ releases from [go.dev/dl](https://go.dev/dl/) also work.

## Install

### Option A — `go install` (users)

```bash
go install github.com/aravinddudam/kubecone/cmd/kubecone@latest
```

The binary is written to `$(go env GOPATH)/bin` (usually `~/go/bin`). Put that directory on `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
kubecone version
```

If `@latest` cannot find a tagged release, install from `main`:

```bash
go install github.com/aravinddudam/kubecone/cmd/kubecone@main
```

### Option B — clone and build (contributors)

```bash
git clone https://github.com/aravinddudam/kubecone.git
cd kubecone
go build -o bin/kubecone ./cmd/kubecone
./bin/kubecone version
```

On Windows:

```powershell
go build -o bin/kubecone.exe ./cmd/kubecone
.\bin\kubecone.exe version
```

Makefile (Unix make):

```bash
make build    # bin/kubecone
make test
make vet
```

## Cluster access

KubeCone uses the same resolution order as kubectl:

1. `--kubeconfig` if you pass it
2. `$KUBECONFIG`
3. `~/.kube/config`
4. in-cluster config (if you ever run it as a pod)

```bash
kubecone investigate deploy/api --kubeconfig ~/.kube/config --context kind-kubecone -n default
```

`-n` / `--namespace` overrides the context default. A bare name (`kubecone investigate payment-api`) is treated as a **Deployment**.

Supported kinds: `pod`, `deploy`/`deployment`, `rs`/`replicaset`, `svc`/`service`, `sts`/`statefulset`, `ds`/`daemonset`, `job`, `cronjob`.

## Quick start

Against a live cluster:

```bash
kubectl get deploy -A
kubecone scan -n banking-dev --unhealthy
kubecone investigate deployment/customer-service -n banking-dev
kubecone investigate deployment/customer-service -n banking-dev -o json
kubecone investigate deployment/customer-service -n banking-dev --quiet --fail
```

Against the bundled kind lab (optional):

```bash
kind create cluster --name kubecone
kubectl apply -f tests/incidents/namespace.yaml
kubectl apply -f tests/incidents/crash-loop/manifest.yaml
kubecone investigate deployment/checkout-worker -n kubecone-lab
kubecone scan -n kubecone-lab --unhealthy
kind delete cluster --name kubecone
```

## Commands

| Command | Alias | Purpose |
| --- | --- | --- |
| `kubecone investigate RESOURCE` | `inv`, `diag` | Collect evidence and diagnose one object |
| `kubecone scan` | `ls` | List workloads and attach a failure code |
| `kubecone explain [CODE]` | | Describe a diagnosis code and the next step |
| `kubecone version` | | Print version, Go runtime, OS/arch |
| `kubecone --help` | | Command and flag reference |
| `kubecone completion bash\|zsh\|fish\|powershell` | | Shell completion |

```bash
# One workload
kubecone investigate deployment/customer-service -n banking-dev
kubecone investigate pod/customer-service-5859cfc476-bbkc6 -n banking-dev
kubecone investigate banking-dev/deploy/fraud-api -o json

# Namespace scan
kubecone scan -n banking-dev
kubecone scan -n banking-dev --unhealthy
kubecone scan -A --kind all
kubecone scan -n kube-system --kind daemonset

# Codes
kubecone explain
kubecone explain IMAGE_PULL
kubecone version --short
```

## Flags

### Global

| Flag | Default | Meaning |
| --- | --- | --- |
| `--kubeconfig` | kubectl rules | path to kubeconfig |
| `--context` | current context | kubeconfig context |
| `-n`, `--namespace` | context namespace | target namespace |
| `-o`, `--output` | `text` | `text` or `json` |
| `-q`, `--quiet` | off | primary diagnosis only |
| `--no-color` | off | disable ANSI color (`NO_COLOR` also works) |
| `--fail` | off | exit `2` when a failure is diagnosed (CI) |
| `--version` | | print version and exit |

### `investigate`

| Flag | Default | Meaning |
| --- | --- | --- |
| `--timeout` | `30s` | collection timeout |
| `--log-tail` | `80` | log lines per container |
| `--events` | `8` | recent events to print (`0` hides them) |
| `--no-graph` | off | omit the evidence graph |
| `--ai` | off | send ranked findings to a provider (stubs in v0.1) |
| `--ai-provider` | none | `openai`, `anthropic`, or `ollama` |

### `scan`

| Flag | Default | Meaning |
| --- | --- | --- |
| `-A`, `--all-namespaces` | off | scan every namespace |
| `--unhealthy` | off | only print workloads with a failure code |
| `--kind` | `deployment` | `deployment`, `statefulset`, `daemonset`, or `all` |
| `--timeout` | `30s` | scan timeout |

`--ai` does nothing useful yet. Providers stay stubbed so the pipeline remains collect → graph → rules → report.

## Output and exit codes

Text output (default) includes the primary diagnosis, related findings, evidence graph, and recent events. Skipped AI is hidden unless you pass `--ai`.

JSON (`-o json`) is the contract for scripts and CI. Compare `primary.code` to the fixtures in `tests/incidents/*/expected.json`.

| Exit code | Meaning |
| --- | --- |
| `0` | success (or no failure when `--fail` is unset) |
| `1` | command or cluster error (bad flags, missing resource, kubeconfig) |
| `2` | a failure was diagnosed **and** `--fail` was set |

Example CI usage:

```bash
kubecone investigate deploy/checkout-worker -n kubecone-lab --quiet --fail
kubecone scan -n kubecone-lab --unhealthy --fail
```

## Diagnosis codes

| Code | Meaning | Typical next step |
| --- | --- | --- |
| `IMAGE_PULL` | image cannot be pulled | fix tag, registry reachability, or credentials (ECR → IAM/IRSA) |
| `FAILED_SCHEDULING` | pod never bound to a node | requests, nodeSelector, taints, PVC |
| `OOM_KILLED` | container exceeded its memory limit | raise the memory limit |
| `PROBE_FAILED` | readiness or liveness probe failing | probe path, port, `initialDelaySeconds` |
| `CRASH_LOOP` | container is restarting | last exit code and logs |
| `HEALTHY` | no rule matched | inspect events/logs, or pass a more specific pod |

`kubecone explain CODE` prints the same catalog.

## Develop and test

```bash
git clone https://github.com/aravinddudam/kubecone.git
cd kubecone
go test ./...
go vet ./...
go build -o bin/kubecone ./cmd/kubecone
```

Live kind tests (needs a cluster):

```bash
go test -tags=kind -timeout 5m ./tests/integration
```

Broken-app fixtures: [tests/incidents/README.md](tests/incidents/README.md). Architecture: [docs/architecture.md](docs/architecture.md).

v0.1 Go module dependencies: `k8s.io/client-go`, `k8s.io/apimachinery`, `k8s.io/api`, `github.com/spf13/cobra`. Prometheus, OpenTelemetry, and controller-runtime are not required.

## Repository layout

```
cmd/kubecone/           CLI entrypoint
internal/cli/           investigate, scan, explain, version
internal/kubernetes/    client-go access layer and kind/name parsing
internal/collector/     pods, events, logs, metrics seam
internal/evidence/      snapshot, graph, findings
internal/analyzer/      oom, crashloop, imagepull, scheduling, probe
internal/engine/        investigate orchestration and namespace scan
internal/reporter/      terminal + JSON
internal/ai/            provider stubs
tests/incidents/        broken-app fixtures for kind
.github/workflows/      CI (test, build, kubecone --help)
```

## License

This project is licensed under the **MIT License**. Copyright (c) 2026 Aravind Dudam.

You may use, copy, modify, merge, publish, distribute, sublicense, and sell copies of the software, provided the copyright notice and permission notice are included in all copies or substantial portions.

The software is provided “as is”, without warranty of any kind. See [LICENSE](LICENSE) for the full text.

# Investigate example

```bash
kubecone investigate deployment/payment-api -n kubecone-lab
```

Expected terminal shape:

```
KubeCone investigation
Target:  kubecone-lab/Deployment/payment-api
Context: kind-kubecone

Primary diagnosis
  [CRITICAL] Container killed because it ran out of memory
  Code:     OOM_KILLED
  Resource: kubecone-lab/Pod/payment-api-xxxx
  payment-api-xxxx/app was OOMKilled (restarts=7, memory limit=24Mi).
  Next:     Raise memory limit from 24Mi to 36Mi (50% headroom after OOMKill).
```

JSON is the contract for tests and later AI:

```bash
kubecone investigate deployment/payment-api -n kubecone-lab -o json
```

The `primary.code` field is what CI compares against `tests/incidents/*/expected.json`.

Scan a namespace first when you do not know which workload is broken. `pods --unhealthy` lists failing pods even when they are not owned by a Deployment:

```bash
kubecone pods -n kubecone-lab --unhealthy
kubecone scan -n kubecone-lab --unhealthy
kubecone explain IMAGE_PULL
```

`--quiet --fail` is the CI shape: print the primary diagnosis and exit 2 when a rule matched.

```bash
kubecone investigate deployment/payment-api -n kubecone-lab --quiet --fail
```

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

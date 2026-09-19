# KubeCone Test Lab

These manifests create broken workloads on purpose. They are original fixtures for
KubeCone, not copies of another project's demo YAMLs.

Apply one scenario at a time against a disposable cluster (kind is the intended lab):

```bash
kind create cluster --name kubecone
kubectl apply -f tests/incidents/namespace.yaml
kubectl apply -f tests/incidents/crash-loop/manifest.yaml
# wait until the pod is CrashLoopBackOff
go run ./cmd/kubecone investigate deployment/checkout-worker -n kubecone-lab -o json
```

| Scenario | Resource | Expected primary code |
| --- | --- | --- |
| oom-killed | `deployment/payment-api` | `OOM_KILLED` |
| crash-loop | `deployment/checkout-worker` | `CRASH_LOOP` |
| image-pull | `deployment/catalog-api` | `IMAGE_PULL` |
| pending-pod | `deployment/batch-importer` | `FAILED_SCHEDULING` |
| failed-readiness | `deployment/storefront` | `PROBE_FAILED` |
| failed-liveness | `deployment/session-cache` | `PROBE_FAILED` or `CRASH_LOOP` |
| pvc | `deployment/ledger-writer` | `FAILED_SCHEDULING` |
| service-routing | `service/payments` | later analyzer |
| dns | `deployment/dns-checker` | later analyzer |
| bad-config | `deployment/config-reader` | `CRASH_LOOP` |

Clean up with `kubectl delete ns kubecone-lab`.

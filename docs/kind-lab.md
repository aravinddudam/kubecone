# Kind lab

kind is the KubeCone test laboratory. It is not a library dependency.

```bash
kind create cluster --name kubecone
kubectl apply -f tests/incidents/namespace.yaml
kubectl apply -f tests/incidents/image-pull/manifest.yaml
go run ./cmd/kubecone investigate deployment/catalog-api -n kubecone-lab -o json
kind delete cluster --name kubecone
```

CI currently runs unit tests only. To exercise a live cluster:

```bash
go test -tags=kind -timeout 5m ./tests/integration
```

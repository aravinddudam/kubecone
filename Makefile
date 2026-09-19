BINARY := bin/kubecone
VERSION ?= 0.2.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null)
MODULE := github.com/aravinddudam/kubecone

.PHONY: build test tidy vet kind-test

build:
	go build -ldflags "-X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT)" -o $(BINARY) ./cmd/kubecone

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

kind-test:
	go test -tags=kind -timeout 5m ./tests/integration

package main

import (
	"os"

	"github.com/aravinddudam/kubecone/internal/cli"
)

func main() {
	if err := cli.New().Execute(); err != nil {
		os.Exit(1)
	}
}

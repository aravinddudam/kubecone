package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/aravinddudam/kubecone/internal/cli"
)

func main() {
	if err := cli.New().Execute(); err != nil {
		if !errors.Is(err, cli.ErrFinding) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(cli.ExitCode(err))
	}
}

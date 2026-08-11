package main

import (
	"fmt"
	"os"

	"github.com/aaronfaby/kalshi-perp-cli/internal/cli"
)

// Set via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	cli.Version = version
	root := cli.NewRoot()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

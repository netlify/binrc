//go:generate rm -rf ./statik
//go:generate gobin -m -run github.com/rakyll/statik -src=./templates

package main

import (
	"fmt"
	"os"

	"github.com/netlify/binrc/cmd"
)

func main() {
	if err := cmd.RootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run command: %v\n", err)
		os.Exit(1)
	}
}

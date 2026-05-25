//go:build !js || !wasm

// Command gutter is the developer CLI: it scaffolds new gutter projects,
// builds them to WebAssembly, serves the result for local development, and
// packages them for deployment.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.7.0"

func main() {
	root := &cobra.Command{
		Use:           "gutter",
		Short:         "Gutter — a declarative Go web framework",
		Long:          styleAccent.Render("Gutter") + " — a declarative Go web framework that compiles to WebAssembly.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		// Bare `gutter` greets with the logo banner, then the usual help.
		Run: func(cmd *cobra.Command, _ []string) {
			printBanner()
			_ = cmd.Help()
		},
	}
	root.AddCommand(newCmd(), runCmd(), buildCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, styleErr.Render("✗")+" "+err.Error())
		os.Exit(1)
	}
}

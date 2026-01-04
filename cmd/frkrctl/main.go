package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "frkrctl",
	Short: "CLI tooling for the Traffic Mirroring Platform operator",
	Long:  `frkrctl provides a developer-friendly interface to the Kubernetes operator via CRDs.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

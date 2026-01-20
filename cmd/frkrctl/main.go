package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "frkrctl",
	Short: "CLI tooling for the frkr operator",
	Long:  `frkrctl provides a developer-friendly interface to the Kubernetes operator via CRDs.`,
}

var outputFormat string

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

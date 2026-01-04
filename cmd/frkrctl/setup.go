package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard",
	Long:  `Interactive setup wizard for initial platform configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSetupWizard()
	},
}

func runSetupWizard() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== frkr Setup Wizard ===")
	fmt.Println()

	// Auth type
	fmt.Println("1. Authentication Configuration")
	fmt.Print("Select auth type (basic/oidc) [basic]: ")
	authType, _ := reader.ReadString('\n')
	authType = strings.TrimSpace(authType)
	if authType == "" {
		authType = "basic"
	}
	fmt.Printf("Selected: %s\n", authType)
	fmt.Println()

	// Data plane
	fmt.Println("2. Data Plane Configuration")
	fmt.Print("Use full stack (includes PostgreSQL-compatible DB and Kafka-compatible broker)? (y/n) [n]: ")
	fullStack, _ := reader.ReadString('\n')
	fullStack = strings.TrimSpace(fullStack)
	if fullStack == "" {
		fullStack = "n"
	}
	fmt.Printf("Selected: %s\n", fullStack)
	fmt.Println()

	// Ingress
	fmt.Println("3. Ingress Configuration")
	fmt.Print("Configure ingress? (y/n) [y]: ")
	ingress, _ := reader.ReadString('\n')
	ingress = strings.TrimSpace(ingress)
	if ingress == "" {
		ingress = "y"
	}
	fmt.Printf("Selected: %s\n", ingress)
	fmt.Println()

	// Summary
	fmt.Println("=== Configuration Summary ===")
	fmt.Printf("Auth Type: %s\n", authType)
	fmt.Printf("Full Stack: %s\n", fullStack)
	fmt.Printf("Ingress: %s\n", ingress)
	fmt.Println()

	fmt.Print("Apply this configuration? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm == "y" {
		fmt.Println("Creating CRDs...")
		// TODO: Create CRDs based on configuration
		fmt.Println("Setup complete!")
	} else {
		fmt.Println("Setup cancelled.")
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

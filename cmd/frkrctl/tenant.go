package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

var tenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Manage tenants",
	Long:  `Create and list tenants via the operator.`,
}

var tenantCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new tenant",
	Long:  `Create a new tenant via the operator (creates FrkrTenant CRD).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Get k8s client
		k8sClient, err := getK8sClient()
		if err != nil {
			return err
		}

		// Get namespace
		ns, err := getNamespace()
		if err != nil {
			return err
		}

		// Create FrkrTenant CRD
		tenant := &frkrv1.FrkrTenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
			},
			Spec: frkrv1.FrkrTenantSpec{
				Name: name,
			},
		}

		if err := k8sClient.Create(context.Background(), tenant); err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}

		fmt.Printf("✅ Tenant request submitted for '%s'\n", name)
		fmt.Println("Waiting for Tenant ID...")

		// Poll for ID
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				fmt.Printf("⚠️  Timed out waiting for tenant ID. Check status with: kubectl get frkrtenant %s -o yaml\n", name)
				return nil
			case <-ticker.C:
				var updatedTenant frkrv1.FrkrTenant
				if err := k8sClient.Get(context.Background(), client.ObjectKey{
					Name:      name,
					Namespace: ns,
				}, &updatedTenant); err == nil {
					if updatedTenant.Status.ID != "" {
						fmt.Printf("\n✅ Tenant ready!\n")
						fmt.Printf("ID: %s\n", updatedTenant.Status.ID)
						return nil
					}
				}
			}
		}
	},
}

var tenantGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get tenant details",
	Long:  `Get tenant details and ID from the operator (FrkrTenant CRD).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Get k8s client
		k8sClient, err := getK8sClient()
		if err != nil {
			return err
		}

		ns, err := getNamespace()
		if err != nil {
			return err
		}

		var tenant frkrv1.FrkrTenant
		if err := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      name,
			Namespace: ns,
		}, &tenant); err != nil {
			return fmt.Errorf("failed to get tenant '%s': %w", name, err)
		}

		if outputFormat == "json" {
			out := map[string]string{
				"id":   tenant.Status.ID,
				"name": tenant.Name,
			}
			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return err
			}
			// If ID is empty, the consumer will see "id": "" which indicates pending.
		} else {
			if tenant.Status.ID == "" {
				fmt.Fprintf(os.Stderr, "⚠️  Tenant '%s' exists but has no ID yet (Operator processing...)\n", name)
			} else {
				fmt.Println(tenant.Status.ID)
			}
		}

		return nil
	},
}

func init() {
	tenantCmd.AddCommand(tenantCreateCmd)
	tenantCmd.AddCommand(tenantGetCmd)
	rootCmd.AddCommand(tenantCmd)
}

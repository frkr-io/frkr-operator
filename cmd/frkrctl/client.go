package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Manage client credentials via operator",
	Long:  `Create and list client credentials via Kubernetes CRDs.`,
}

var clientCreateCmd = &cobra.Command{
	Use:   "create [client-id]",
	Short: "Create a new client credential",
	Long:  `Create a new OAuth client credential via the operator (creates FrkrClient CRD).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID := args[0]
		tenantID, _ := cmd.Flags().GetString("tenant-id")
		streamID, _ := cmd.Flags().GetString("stream-id")
		secret, _ := cmd.Flags().GetString("secret")

		if tenantID == "" {
			return fmt.Errorf("--tenant-id is required")
		}

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

		// Create FrkrClient CRD
		crdName := fmt.Sprintf("%s-client", clientID)
		crd := &frkrv1.FrkrClient{
			ObjectMeta: metav1.ObjectMeta{
				Name:      crdName,
				Namespace: ns,
			},
			Spec: frkrv1.FrkrClientSpec{
				TenantID: tenantID,
				ClientID: clientID,
				StreamID: streamID,
				Secret:   secret,
			},
		}

		if err := k8sClient.Create(context.Background(), crd); err != nil {
			return fmt.Errorf("failed to create client CRD: %w", err)
		}

		fmt.Printf("✅ Client request submitted for '%s'\n", clientID)
		fmt.Println("Waiting for secret generation...")

		// Poll for Secret
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		secretName := fmt.Sprintf("frkr-client-%s", crdName)

		for {
			select {
			case <-timeout:
				fmt.Printf("⚠️  Timed out waiting for secret. Check status with: kubectl get frkrclient %s -o yaml\n", crdName)
				return nil
			case <-ticker.C:
				var secret corev1.Secret
				if err := k8sClient.Get(context.Background(), client.ObjectKey{
					Name:      secretName,
					Namespace: ns,
				}, &secret); err == nil {
					clientSecret := string(secret.Data["clientSecret"])
					if clientSecret != "" {
						fmt.Printf("\n✅ Client Credential Ready!\n")
						fmt.Printf("ClientID:     %s\n", clientID)
						fmt.Printf("ClientSecret: %s\n", clientSecret)
						fmt.Println("\nSave this secret! It is stored in a Kubernetes Secret but displayed here for convenience.")
						return nil
					}
				}
			}
		}
	},
}

var clientListCmd = &cobra.Command{
	Use:   "list",
	Short: "List client credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		k8sClient, err := getK8sClient()
		if err != nil {
			return err
		}

		ns, err := getNamespace()
		if err != nil {
			return err
		}

		var list frkrv1.FrkrClientList
		if err := k8sClient.List(context.Background(), &list, client.InNamespace(ns)); err != nil {
			return fmt.Errorf("failed to list clients: %w", err)
		}

		if len(list.Items) == 0 {
			fmt.Println("No clients found.")
			return nil
		}

		fmt.Printf("%-20s %-20s %-36s %-10s\n", "NAME", "CLIENT ID", "TENANT ID", "STATUS")
		for _, item := range list.Items {
			fmt.Printf("%-20s %-20s %-36s %-10s\n", item.Name, item.Spec.ClientID, item.Spec.TenantID, item.Status.Phase)
		}
		return nil
	},
}

func init() {
	clientCreateCmd.Flags().String("tenant-id", "", "Tenant ID (required)")
	clientCreateCmd.Flags().String("stream-id", "", "Stream ID to scope to (optional)")
	clientCreateCmd.Flags().String("secret", "", "Optional custom secret")

	clientCmd.AddCommand(clientCreateCmd)
	clientCmd.AddCommand(clientListCmd)

	rootCmd.AddCommand(clientCmd)
}

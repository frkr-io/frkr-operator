package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `Create, list, reset passwords, and delete users via the operator.`,
}

var userCreateCmd = &cobra.Command{
	Use:   "create [username]",
	Short: "Create a new user",
	Long:  `Create a new user via the operator (creates FrkrUser CRD).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]
		tenantID, _ := cmd.Flags().GetString("tenant-id")

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

		// Create FrkrUser CRD
		user := &frkrv1.FrkrUser{
			ObjectMeta: metav1.ObjectMeta{
				Name:      username,
				Namespace: ns,
			},
			Spec: frkrv1.FrkrUserSpec{
				Username: username,
				TenantID: tenantID,
			},
		}

		if err := k8sClient.Create(context.Background(), user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		if outputFormat != "json" {
			fmt.Printf("✅ User %s created successfully\n", username)
			fmt.Println("Waiting for password generation...")
		}

		// Poll for secret
		timeoutSeconds, _ := cmd.Flags().GetInt("timeout")
		timeout := time.After(time.Duration(timeoutSeconds) * time.Second)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				if outputFormat == "json" {
					return fmt.Errorf("timed out waiting for password (%ds)", timeoutSeconds)
				}
				fmt.Printf("⚠️  Timed out waiting for password (%ds). Check status with: kubectl get secret frkr-user-%s -o yaml\n", timeoutSeconds, username)
				return nil
			case <-ticker.C:
				var secret corev1.Secret
				if err := k8sClient.Get(context.Background(), client.ObjectKey{
					Name:      fmt.Sprintf("frkr-user-%s", username),
					Namespace: ns,
				}, &secret); err == nil {
					if pass, ok := secret.Data["password"]; ok {
						if outputFormat == "json" {
							// JSON Output
							out := map[string]string{
								"username":  username,
								"password":  string(pass),
								"tenant_id": tenantID,
								"status":    "active",
							}
							if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
								// Fallback to error
								return err
							}
						} else {
							fmt.Printf("\n🔑 Password: %s\n\n", string(pass))
							fmt.Println("Save this password! It will not be shown again.")
						}
						return nil
					}
				}
			}
		}
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Long:  `List all users managed by the operator.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		k8sClient, err := getK8sClient()
		if err != nil {
			return err
		}

		ns, err := getNamespace()
		if err != nil {
			return err
		}

		var userList frkrv1.FrkrUserList
		if err := k8sClient.List(context.Background(), &userList, client.InNamespace(ns)); err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}

		if len(userList.Items) == 0 {
			fmt.Println("No users found")
			return nil
		}

		fmt.Printf("%-20s %-36s %-10s\n", "NAME", "TENANT ID", "PHASE")
		fmt.Println("--------------------------------------------------------------------------------")
		for _, user := range userList.Items {
			fmt.Printf("%-20s %-36s %-10s\n",
				user.Spec.Username,
				user.Spec.TenantID,
				user.Status.Phase)
		}

		return nil
	},
}

var userResetPasswordCmd = &cobra.Command{
	Use:   "reset-password [username]",
	Short: "Reset a user's password",
	Long:  `Reset a user's password (updates FrkrUser CRD to trigger reset).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]

		k8sClient, err := getK8sClient()
		if err != nil {
			return err
		}

		ns, err := getNamespace()
		if err != nil {
			return err
		}

		var user frkrv1.FrkrUser
		if err := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      username,
			Namespace: ns,
		}, &user); err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Clear password to trigger regeneration
		user.Spec.Password = ""
		if err := k8sClient.Update(context.Background(), &user); err != nil {
			return fmt.Errorf("failed to reset password: %w", err)
		}

		fmt.Printf("✅ Password reset for user %s\n", username)
		fmt.Printf("Check new password with: kubectl get frkruser %s -o jsonpath='{.status.password}'\n", username)
		return nil
	},
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete [username]",
	Short: "Delete a user",
	Long:  `Delete a user (deletes FrkrUser CRD).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]

		k8sClient, err := getK8sClient()
		if err != nil {
			return err
		}

		ns, err := getNamespace()
		if err != nil {
			return err
		}

		var user frkrv1.FrkrUser
		if err := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      username,
			Namespace: ns,
		}, &user); err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if err := k8sClient.Delete(context.Background(), &user); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		fmt.Printf("✅ User %s deleted\n", username)
		return nil
	},
}

func init() {
	userCreateCmd.Flags().String("tenant-id", "", "Tenant ID (required)")
	userCreateCmd.Flags().Int("timeout", 90, "Timeout in seconds to wait for password generation")
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userResetPasswordCmd)
	userCmd.AddCommand(userDeleteCmd)
	rootCmd.AddCommand(userCmd)
}

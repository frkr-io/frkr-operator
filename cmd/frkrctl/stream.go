package main

import (
	"context"
	"fmt"

	"github.com/frkr-io/frkr-common/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Manage streams",
	Long:  `Create, list, and manage streams via the operator (creates FrkrStream CRDs).`,
}

var streamCreateCmd = &cobra.Command{
	Use:   "create [stream-name]",
	Short: "Create a new stream",
	Long:  `Create a new stream via the operator (creates FrkrStream CRD).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		streamName := args[0]
		tenantID, _ := cmd.Flags().GetString("tenant-id")
		description, _ := cmd.Flags().GetString("description")
		retentionDays, _ := cmd.Flags().GetInt("retention-days")

		if tenantID == "" {
			return fmt.Errorf("--tenant-id is required")
		}

		// Use shared stream name validation
		if err := util.ValidateStreamName(streamName); err != nil {
			return err
		}

		// Normalize and validate retention days
		normalizedDays, err := util.NormalizeRetentionDays(retentionDays)
		if err != nil {
			return err
		}
		retentionDays = normalizedDays

		// Get k8s client
		k8sClient, err := getK8sClient()
		if err != nil {
			return err
		}

		// Create FrkrStream CRD
		stream := &frkrv1.FrkrStream{
			ObjectMeta: metav1.ObjectMeta{
				Name:      streamName,
				Namespace: getNamespace(),
			},
			Spec: frkrv1.FrkrStreamSpec{
				TenantID:      tenantID,
				Name:          streamName,
				Description:   description,
				RetentionDays: retentionDays,
			},
		}

		if err := k8sClient.Create(context.Background(), stream); err != nil {
			return fmt.Errorf("failed to create stream: %w", err)
		}

		fmt.Printf("✅ Stream %s created successfully\n", streamName)
		fmt.Printf("Check status with: kubectl get frkrstream %s -o yaml\n", streamName)
		return nil
	},
}

var streamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all streams",
	Long:  `List all streams managed by the operator.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		k8sClient, err := getK8sClient()
		if err != nil {
			return err
		}

		var streamList frkrv1.FrkrStreamList
		if err := k8sClient.List(context.Background(), &streamList, client.InNamespace(getNamespace())); err != nil {
			return fmt.Errorf("failed to list streams: %w", err)
		}

		if len(streamList.Items) == 0 {
			fmt.Println("No streams found")
			return nil
		}

		fmt.Printf("%-20s %-36s %-30s %-10s\n", "NAME", "TENANT ID", "TOPIC", "PHASE")
		fmt.Println("--------------------------------------------------------------------------------------------------------")
		for _, stream := range streamList.Items {
			fmt.Printf("%-20s %-36s %-30s %-10s\n",
				stream.Spec.Name,
				stream.Spec.TenantID,
				stream.Status.Topic,
				stream.Status.Phase)
		}

		return nil
	},
}

func init() {
	streamCreateCmd.Flags().String("tenant-id", "", "Tenant ID (required)")
	streamCreateCmd.Flags().String("description", "", "Stream description")
	streamCreateCmd.Flags().Int("retention-days", 7, "Retention period in days (default: 7)")

	streamCmd.AddCommand(streamCreateCmd)
	streamCmd.AddCommand(streamListCmd)
	rootCmd.AddCommand(streamCmd)
}

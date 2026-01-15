package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Manage Prometheus metrics collection",
	Long:  `Enable, disable, or check status of Prometheus ServiceMonitor for frkr gateways.`,
}

var metricsEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable Prometheus metrics collection",
	Long:  `Creates a ServiceMonitor CRD to enable Prometheus scraping of frkr gateways.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, _ := cmd.Flags().GetString("namespace")
		return enableMetrics(namespace)
	},
}

var metricsDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable Prometheus metrics collection",
	Long:  `Removes the ServiceMonitor CRD to disable Prometheus scraping.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, _ := cmd.Flags().GetString("namespace")
		return disableMetrics(namespace)
	},
}

var metricsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check metrics collection status",
	Long:  `Check if ServiceMonitor is deployed and Prometheus is scraping frkr gateways.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, _ := cmd.Flags().GetString("namespace")
		return metricsStatus(namespace)
	},
}

// getDynamicClient creates a dynamic Kubernetes client
func getDynamicClient() (dynamic.Interface, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s config: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return dynamicClient, nil
}

func enableMetrics(namespace string) error {
	client, err := getDynamicClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Check if ServiceMonitor CRD exists
	if !serviceMonitorCRDExists(client) {
		return fmt.Errorf("ServiceMonitor CRD not found. Please install Prometheus Operator first.\n" +
			"See: https://prometheus-operator.dev/docs/getting-started/installation/")
	}

	// Create ServiceMonitor
	sm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "monitoring.coreos.com/v1",
			"kind":       "ServiceMonitor",
			"metadata": map[string]interface{}{
				"name":      "frkr-gateways",
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app.kubernetes.io/part-of": "frkr",
				},
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.kubernetes.io/part-of": "frkr",
					},
				},
				"endpoints": []interface{}{
					map[string]interface{}{
						"port":     "metrics",
						"path":     "/metrics",
						"interval": "30s",
					},
				},
				"namespaceSelector": map[string]interface{}{
					"matchNames": []interface{}{namespace},
				},
			},
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "servicemonitors",
	}

	_, err = client.Resource(gvr).Namespace(namespace).Create(context.Background(), sm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ServiceMonitor: %w", err)
	}

	fmt.Printf("✅ ServiceMonitor 'frkr-gateways' created in namespace '%s'\n", namespace)
	fmt.Println("   Prometheus will begin scraping /metrics endpoints shortly.")
	return nil
}

func disableMetrics(namespace string) error {
	client, err := getDynamicClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "servicemonitors",
	}

	err = client.Resource(gvr).Namespace(namespace).Delete(context.Background(), "frkr-gateways", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete ServiceMonitor: %w", err)
	}

	fmt.Printf("✅ ServiceMonitor 'frkr-gateways' deleted from namespace '%s'\n", namespace)
	return nil
}

func metricsStatus(namespace string) error {
	client, err := getDynamicClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Check CRD
	if !serviceMonitorCRDExists(client) {
		fmt.Println("❌ ServiceMonitor CRD not installed")
		fmt.Println("   Prometheus Operator is required for metrics collection.")
		return nil
	}
	fmt.Println("✅ ServiceMonitor CRD available")

	// Check if our ServiceMonitor exists
	gvr := schema.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "servicemonitors",
	}

	sm, err := client.Resource(gvr).Namespace(namespace).Get(context.Background(), "frkr-gateways", metav1.GetOptions{})
	if err != nil {
		fmt.Printf("❌ ServiceMonitor 'frkr-gateways' not found in namespace '%s'\n", namespace)
		fmt.Println("   Run 'frkrctl metrics enable' to create it.")
		return nil
	}

	fmt.Printf("✅ ServiceMonitor 'frkr-gateways' deployed in namespace '%s'\n", namespace)

	// Show creation time
	creationTime := sm.GetCreationTimestamp()
	fmt.Printf("   Created: %s\n", creationTime.Format("2006-01-02 15:04:05"))

	return nil
}

func serviceMonitorCRDExists(client dynamic.Interface) bool {
	gvr := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	_, err := client.Resource(gvr).Get(context.Background(), "servicemonitors.monitoring.coreos.com", metav1.GetOptions{})
	return err == nil
}

func init() {
	metricsCmd.PersistentFlags().StringP("namespace", "n", "frkr", "Kubernetes namespace")

	metricsCmd.AddCommand(metricsEnableCmd)
	metricsCmd.AddCommand(metricsDisableCmd)
	metricsCmd.AddCommand(metricsStatusCmd)

	rootCmd.AddCommand(metricsCmd)
}

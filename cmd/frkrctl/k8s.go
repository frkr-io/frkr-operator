package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

// getK8sClient creates and returns a Kubernetes client configured for frkr CRDs.
// This is a shared utility to avoid code duplication across frkrctl commands.
//
// TODO: Consider moving this to frkr-common/k8s/client.go if frkr-tools needs
// to interact with the operator (e.g., for CRD-based operations). Currently
// this is operator-specific as it requires frkr CRD types.
func getK8sClient() (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s config: %w", err)
	}

	scheme := runtime.NewScheme()
	frkrv1.AddToScheme(scheme)

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return k8sClient, nil
}

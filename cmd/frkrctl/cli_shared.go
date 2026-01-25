package main

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

func getNamespace() (string, error) {
	// 1. Env var has highest precedence (after flags, but flags handled by cobra usually)
	ns := os.Getenv("FRKR_NAMESPACE")
	if ns != "" {
		return ns, nil
	}

	// 2. Try to get from kubeconfig current context
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	namespace, _, err := kubeConfig.Namespace()
	if err == nil && namespace != "" {
		return namespace, nil
	}

	// 3. Fail if no namespace found
	return "", fmt.Errorf("namespace not specified (use --namespace, context, or FRKR_NAMESPACE)")
}

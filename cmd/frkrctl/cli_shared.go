package main

import "os"

func getNamespace() string {
	ns := os.Getenv("FRKR_NAMESPACE")
	if ns == "" {
		return "default"
	}
	return ns
}

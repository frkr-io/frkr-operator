package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// SetupControllers sets up all controllers
func SetupControllers(mgr manager.Manager) error {
	// Setup User controller
	if err := (&UserReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	// Setup AuthConfig controller
	if err := (&AuthConfigReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	// Setup DataPlane controller
	if err := (&DataPlaneReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	// Setup Init controller
	if err := (&InitReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}


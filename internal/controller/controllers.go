package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/frkr-io/frkr-operator/internal/infra"
)

// SetupControllers sets up all controllers
func SetupControllers(mgr manager.Manager) error {
	setupLog := log.Log.WithName("setup")

	// Get infrastructure config
	config, err := infra.GetConfigFromEnv()
	if err != nil {
		setupLog.Error(err, "unable to get infrastructure config")
		// Continue anyway, controllers will handle nil connections
	}

	var db *infra.DB
	if config.DatabaseURL != "" {
		db, err = infra.NewDB(config.DatabaseURL)
		if err != nil {
			setupLog.Error(err, "unable to connect to database")
		}
	}

	var kafkaAdmin *infra.KafkaAdmin
	if config.BrokerURL != "" {
		kafkaAdmin = infra.NewKafkaAdmin(config.BrokerURL)
	}

	// Setup User controller
	if err := (&UserReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		DB:     db,
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

	// Setup Stream controller
	if err := (&StreamReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		DB:         db,
		KafkaAdmin: kafkaAdmin,
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}

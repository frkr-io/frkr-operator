package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

// DataPlaneReconciler reconciles a FrkrDataPlane object
type DataPlaneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=frkr.io,resources=frkrdatapanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frkr.io,resources=frkrdatapanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frkr.io,resources=frkrdatapanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *DataPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var dataPlane frkrv1.FrkrDataPlane
	if err := r.Get(ctx, req.NamespacedName, &dataPlane); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Test Postgres connectivity (with warning on error, but still apply)
	postgresConnected := r.testPostgresConnection(ctx, &dataPlane)
	if !postgresConnected {
		logger.Info("warning: Postgres connection test failed", "host", dataPlane.Spec.PostgresConfig.Host)
		dataPlane.Status.Warnings = append(dataPlane.Status.Warnings, "Postgres connectivity test failed")
	}

	// Test Redpanda connectivity (with warning on error, but still apply)
	redpandaConnected := r.testRedpandaConnection(ctx, &dataPlane)
	if !redpandaConnected {
		logger.Info("warning: Redpanda connection test failed", "brokers", dataPlane.Spec.RedpandaConfig.Brokers)
		dataPlane.Status.Warnings = append(dataPlane.Status.Warnings, "Redpanda connectivity test failed")
	}

	// Update status
	dataPlane.Status.PostgresConnected = postgresConnected
	dataPlane.Status.RedpandaConnected = redpandaConnected
	dataPlane.Status.Phase = "Active"

	if err := r.Status().Update(ctx, &dataPlane); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconciled data plane")
	return ctrl.Result{}, nil
}

func (r *DataPlaneReconciler) testPostgresConnection(ctx context.Context, dataPlane *frkrv1.FrkrDataPlane) bool {
	// TODO: Implement actual connection test
	// For now, return true (always apply changes, just warn)
	return true
}

func (r *DataPlaneReconciler) testRedpandaConnection(ctx context.Context, dataPlane *frkrv1.FrkrDataPlane) bool {
	// TODO: Implement actual connection test
	// For now, return true (always apply changes, just warn)
	return true
}

// SetupWithManager sets up the controller with the Manager
func (r *DataPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frkrv1.FrkrDataPlane{}).
		Complete(r)
}


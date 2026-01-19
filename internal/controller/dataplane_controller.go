package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

//+kubebuilder:rbac:groups=frkr.io,resources=frkrdataplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frkr.io,resources=frkrdataplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frkr.io,resources=frkrdataplanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *DataPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var dataPlane frkrv1.FrkrDataPlane
	if err := r.Get(ctx, req.NamespacedName, &dataPlane); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Clear previous warnings and rebuild based on current state
	var warnings []string

	// Test Postgres connectivity (with warning on error, but still apply)
	postgresConnected := r.testPostgresConnection(ctx, &dataPlane)
	if !postgresConnected {
		logger.Info("warning: Postgres connection test failed", "host", dataPlane.Spec.PostgresConfig.Host)
		warnings = append(warnings, "Postgres connectivity test failed")
	}

	// Test Kafka-compatible broker connectivity (with warning on error, but still apply)
	brokerConnected := r.testBrokerConnection(ctx, &dataPlane)
	if !brokerConnected {
		logger.Info("warning: broker connection test failed", "brokers", dataPlane.Spec.BrokerConfig.Brokers)
		warnings = append(warnings, "Kafka-compatible broker connectivity test failed")
	}

	// Update status
	dataPlane.Status.PostgresConnected = postgresConnected
	dataPlane.Status.BrokerConnected = brokerConnected
	dataPlane.Status.Warnings = warnings
	dataPlane.Status.Phase = "Active"
	
	// Determine readiness
	ready := postgresConnected && brokerConnected
	status := metav1.ConditionFalse
	reason := "ComponentsUnhealthy"
	msg := "One or more components are not reachable"
	if ready {
		status = metav1.ConditionTrue
		reason = "ComponentsHealthy"
		msg = "All data plane components are connected and healthy"
	}
	
	meta.SetStatusCondition(&dataPlane.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  status,
		Reason:  reason,
		Message: msg,
	})

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

func (r *DataPlaneReconciler) testBrokerConnection(ctx context.Context, dataPlane *frkrv1.FrkrDataPlane) bool {
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

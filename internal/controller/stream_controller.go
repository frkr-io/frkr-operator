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
	// TODO: Import db utilities when operator has DB connection from FrkrDataPlane
)

// StreamReconciler reconciles a FrkrStream object
type StreamReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// TODO: Add DB connection from FrkrDataPlane config
}

//+kubebuilder:rbac:groups=frkr.io,resources=frkrstreams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frkr.io,resources=frkrstreams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frkr.io,resources=frkrstreams/finalizers,verbs=update
//+kubebuilder:rbac:groups=frkr.io,resources=frkrdataplanes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *StreamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var stream frkrv1.FrkrStream
	if err := r.Get(ctx, req.NamespacedName, &stream); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: Get database connection from FrkrDataPlane
	// For now, this is a placeholder - full implementation needs DB connection
	// from operator's data plane configuration

	// Update status
	stream.Status.Phase = "Pending"
	meta.SetStatusCondition(&stream.Status.Conditions, metav1.Condition{
		Type:    "StreamCreated",
		Status:  metav1.ConditionFalse,
		Reason:  "NotImplemented",
		Message: "Stream controller needs database connection from FrkrDataPlane",
	})

	if err := r.Status().Update(ctx, &stream); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconciled stream", "name", stream.Spec.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *StreamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frkrv1.FrkrStream{}).
		Complete(r)
}

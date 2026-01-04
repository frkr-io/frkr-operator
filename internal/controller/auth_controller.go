package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

// AuthConfigReconciler reconciles a FrkrAuthConfig object
type AuthConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=frkr.io,resources=frkrauthconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frkr.io,resources=frkrauthconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frkr.io,resources=frkrauthconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups=frkr.io,resources=frkrusers,verbs=list;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *AuthConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var authConfig frkrv1.FrkrAuthConfig
	if err := r.Get(ctx, req.NamespacedName, &authConfig); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if auth type changed
	previousType := authConfig.Status.PreviousType
	if previousType != "" && previousType != authConfig.Spec.Type {
		// Auth type changed - delete all basic auth users if switching away from basic
		if previousType == frkrv1.AuthTypeBasic {
			logger.Info("switching away from basic auth, deleting all basic auth users")
			var userList frkrv1.FrkrUserList
			if err := r.List(ctx, &userList); err == nil {
				for _, user := range userList.Items {
					if err := r.Delete(ctx, &user); err != nil {
						logger.Error(err, "failed to delete user", "username", user.Spec.Username)
					}
				}
			}
		}
	}

	// Update status
	authConfig.Status.PreviousType = authConfig.Spec.Type
	authConfig.Status.Phase = "Active"

	if err := r.Status().Update(ctx, &authConfig); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconciled auth config", "type", authConfig.Spec.Type)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *AuthConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frkrv1.FrkrAuthConfig{}).
		Complete(r)
}

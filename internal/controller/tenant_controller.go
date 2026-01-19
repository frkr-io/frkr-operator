package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
	"github.com/frkr-io/frkr-operator/internal/infra"
)

// TenantReconciler reconciles a FrkrTenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	DB     *infra.DB
}

//+kubebuilder:rbac:groups=frkr.io,resources=frkrtenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frkr.io,resources=frkrtenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frkr.io,resources=frkrtenants/finalizers,verbs=update

func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var tenant frkrv1.FrkrTenant
	if err := r.Get(ctx, req.NamespacedName, &tenant); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Name to use for the tenant (default to CR name if not specified)
	tenantName := tenant.Spec.Name
	if tenantName == "" {
		tenantName = tenant.Name
	}

	log.Info("reconciling tenant", "name", tenantName)

	// Create or Get Tenant in DB
	tenantID, err := r.DB.EnsureTenant(tenantName)
	if err != nil {
		log.Error(err, "failed to ensure tenant")
		return ctrl.Result{}, err
	}

	// Update Status
	if tenant.Status.ID != tenantID || tenant.Status.Phase != "Ready" {
		tenant.Status.ID = tenantID
		tenant.Status.Phase = "Ready"
		if err := r.Status().Update(ctx, &tenant); err != nil {
			log.Error(err, "failed to update tenant status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frkrv1.FrkrTenant{}).
		Complete(r)
}

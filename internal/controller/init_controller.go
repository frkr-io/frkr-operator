package controller

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
	"github.com/frkr-io/frkr-common/migrate"
)

// InitReconciler reconciles a FrkrInit object
type InitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=frkr.io,resources=frkrinits,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frkr.io,resources=frkrinits/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frkr.io,resources=frkrinits/finalizers,verbs=update
//+kubebuilder:rbac:groups=frkr.io,resources=frkrdatapanes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *InitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var init frkrv1.FrkrInit
	if err := r.Get(ctx, req.NamespacedName, &init); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get database URL
	dbURL := init.Spec.DatabaseURL
	if dbURL == "" {
		// Get from FrkrDataPlane
		var dataPlaneList frkrv1.FrkrDataPlaneList
		if err := r.List(ctx, &dataPlaneList); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get data plane config: %w", err)
		}
		if len(dataPlaneList.Items) == 0 {
			return ctrl.Result{}, fmt.Errorf("no data plane configuration found")
		}
		dataPlane := dataPlaneList.Items[0]
		// Build connection string from data plane config
		dbURL = fmt.Sprintf("cockroachdb://%s@%s:%d/%s?sslmode=%s",
			dataPlane.Spec.PostgresConfig.User,
			dataPlane.Spec.PostgresConfig.Host,
			dataPlane.Spec.PostgresConfig.Port,
			dataPlane.Spec.PostgresConfig.Database,
			dataPlane.Spec.PostgresConfig.SSLMode,
		)
	}

	// Get migrations path
	migrationsPath := init.Spec.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "/migrations" // Default path
	}

	// Run migrations
	if err := migrate.RunMigrations(dbURL, migrationsPath); err != nil {
		logger.Error(err, "failed to run migrations")
		init.Status.Phase = "Failed"
		init.Status.Conditions = append(init.Status.Conditions, metav1.Condition{
			Type:    "MigrationsFailed",
			Status:  metav1.ConditionFalse,
			Reason:  "MigrationError",
			Message: err.Error(),
		})
		if err := r.Status().Update(ctx, &init); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// Get version
	version, dirty, err := migrate.GetVersion(dbURL, migrationsPath)
	if err != nil {
		logger.Error(err, "failed to get migration version")
	} else {
		init.Status.Version = version
		init.Status.Dirty = dirty
	}

	// Update status
	init.Status.Phase = "Initialized"
	init.Status.Conditions = append(init.Status.Conditions, metav1.Condition{
		Type:    "MigrationsComplete",
		Status:  metav1.ConditionTrue,
		Reason:  "Success",
		Message: "Database migrations completed successfully",
	})

	if err := r.Status().Update(ctx, &init); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconciled init", "version", version)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *InitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frkrv1.FrkrInit{}).
		Complete(r)
}


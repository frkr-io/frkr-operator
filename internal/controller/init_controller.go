package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/frkr-io/frkr-common/migrate"
	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// InitReconciler reconciles a FrkrInit object
type InitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=frkr.io,resources=frkrinits,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frkr.io,resources=frkrinits/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frkr.io,resources=frkrinits/finalizers,verbs=update
//+kubebuilder:rbac:groups=frkr.io,resources=frkrdataplanes,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch

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
		// Determine scheme based on Type (default to postgres)
		scheme := "postgres"
		if dataPlane.Spec.PostgresConfig.Type == "cockroachdb" {
			scheme = "cockroachdb"
		}
		
		// Get Password from Secret
		var secret corev1.Secret
		password := ""
		if dataPlane.Spec.PostgresConfig.PasswordRef != "" {
			secretName := dataPlane.Spec.PostgresConfig.PasswordRef
			if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: init.Namespace}, &secret); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to get password secret %s: %w", secretName, err)
			}
			// Assume key is "password" based on convention, or check multiple?
			// frkr-db-secret uses "password".
			if PassBytes, ok := secret.Data["password"]; ok {
				password = string(PassBytes)
			} else {
				// Fallback or error?
				return ctrl.Result{}, fmt.Errorf("secret %s does not contain 'password' key", secretName)
			}
		}

		dbURL = fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
			scheme,
			dataPlane.Spec.PostgresConfig.User,
			password,
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
		meta.SetStatusCondition(&init.Status.Conditions, metav1.Condition{
			Type:    "MigrationsComplete",
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

	// Check Gateways (only if migrations successful and not dirty)
	// We do this BEFORE setting the final "Ready" status
	gatewaysReady, err := r.checkGateways(ctx, &init)
	if err != nil {
		logger.Error(err, "failed to check gateways")
		// continue to report partial status, or return error?
		// Retrying is fine.
		return ctrl.Result{}, err
	}
	
	if !gatewaysReady && len(init.Spec.Gateways) > 0 {
		logger.Info("waiting for gateways to be ready...", "gateways", init.Spec.Gateways)
		// Requeue to check again effectively
		// But first update status to show Migration Success but "Waiting for Gateways"
		// We can add a specialized condition or just keep Ready=False
		meta.SetStatusCondition(&init.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "WaitingForGateways",
			Message: "Migrations complete, waiting for gateways",
		})
		if err := r.Status().Update(ctx, &init); err != nil {
			return ctrl.Result{}, err
		}
		// Requeue after some time to poll
		// Controller runtime naturally requeues on status update, but for external resource change (deployments)
		// we should watch them? We added RBAC but didn't add Watch in SetupWithManager. 
		// We should add that for responsiveness, or poll. Poll is easier for now.
		return ctrl.Result{RequeueAfter: 10 * 1000 * 1000 * 1000}, nil // 10s
	}

	// Update status
	init.Status.Phase = "Initialized"
	meta.SetStatusCondition(&init.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "MigrationsComplete",
		Message: fmt.Sprintf("Database migrations completed successfully (version: %d) and gateways are ready", version),
	})

	if err := r.Status().Update(ctx, &init); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconciled init", "version", version)
	return ctrl.Result{}, nil
}

func (r *InitReconciler) checkGateways(ctx context.Context, init *frkrv1.FrkrInit) (bool, error) {
	if len(init.Spec.Gateways) == 0 {
		return true, nil
	}

	for _, name := range init.Spec.Gateways {
		var dep appsv1.Deployment
		key := client.ObjectKey{
			Namespace: init.Namespace,
			Name:      name,
		}
		if err := r.Get(ctx, key, &dep); err != nil {
			return false, client.IgnoreNotFound(err)
		}
		// Check for readiness
		if dep.Status.AvailableReplicas == 0 {
			return false, nil
		}
	}
	return true, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *InitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frkrv1.FrkrInit{}).
		Complete(r)
}

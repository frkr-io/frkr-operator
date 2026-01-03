package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

// UserReconciler reconciles a FrkrUser object
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=frkr.io,resources=frkrusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frkr.io,resources=frkrusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frkr.io,resources=frkrusers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var user frkrv1.FrkrUser
	if err := r.Get(ctx, req.NamespacedName, &user); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Generate password if not provided
	password := user.Spec.Password
	if password == "" {
		// Generate random password
		passwordBytes := make([]byte, 32)
		if _, err := rand.Read(passwordBytes); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to generate password: %w", err)
		}
		password = base64.URLEncoding.EncodeToString(passwordBytes)
		user.Status.PasswordGenerated = true
	}

	// Set password in status (one-time display only)
	if user.Status.Password == "" {
		user.Status.Password = password
		user.Status.Phase = "Active"
	}

	// Create Kubernetes secret for credentials
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("frkr-user-%s", user.Spec.Username),
			Namespace: req.Namespace,
		},
		Data: map[string][]byte{
			"username": []byte(user.Spec.Username),
			"password": []byte(password),
		},
	}

	if err := ctrl.SetControllerReference(&user, secret, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, secret); err != nil {
		logger.Error(err, "failed to create secret")
		// Continue even if secret creation fails (may already exist)
	}

	// Update status
	if err := r.Status().Update(ctx, &user); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconciled user", "username", user.Spec.Username)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frkrv1.FrkrUser{}).
		Complete(r)
}


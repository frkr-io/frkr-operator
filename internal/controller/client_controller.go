package controller

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/frkr-io/frkr-common/util"
	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
	"github.com/frkr-io/frkr-operator/internal/infra"
)

// ClientReconciler reconciles a FrkrClient object
type ClientReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	DB     *infra.DB
}

//+kubebuilder:rbac:groups=frkr.io,resources=frkrclients,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frkr.io,resources=frkrclients/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frkr.io,resources=frkrclients/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *ClientReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var crd frkrv1.FrkrClient
	if err := r.Get(ctx, req.NamespacedName, &crd); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Secret handling
	clientSecret := crd.Spec.Secret
	if clientSecret == "" {
		// Check if a secret already exists in Kubernetes for this client.
		// If not, auto-generate a new one and ensure it persists in a K8s Secret.

		secretName := fmt.Sprintf("frkr-client-%s", crd.Name)
		var existingSecret corev1.Secret
		err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: crd.Namespace}, &existingSecret)
		if err == nil {
			// Found existing secret, check if it has the data
			if val, ok := existingSecret.Data["clientSecret"]; ok {
				clientSecret = string(val)
			}
		}

		if clientSecret == "" {
			var err error
			clientSecret, err = util.GeneratePassword()
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to generate secret: %w", err)
			}
			crd.Status.SecretGenerated = true
		}
	}

	// Persist to DB
	if r.DB != nil {
		// Resolve StreamID.
		// Currently, we expect the Spec.StreamID to be a valid UUID if provided.
		// The infra layer handles the database interaction to ensure the client exists.

		var streamID *string
		if crd.Spec.StreamID != "" {
			s := crd.Spec.StreamID
			streamID = &s
		}

		dbClient, err := r.DB.EnsureClient(crd.Spec.TenantID, crd.Spec.ClientID, clientSecret, streamID)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				// Dependency missing (Tenant/Stream)
				log.Error(err, "dependency missing")
				return ctrl.Result{RequeueAfter: 30}, err
			}
			log.Error(err, "failed to ensure client in db")
			return ctrl.Result{}, err
		}

		crd.Status.ID = dbClient.ID
		crd.Status.Phase = "Ready"
	}

	// Create/Update Kubernetes Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("frkr-client-%s", crd.Name),
			Namespace: req.Namespace,
		},
		Data: map[string][]byte{
			"clientId":     []byte(crd.Spec.ClientID),
			"clientSecret": []byte(clientSecret),
		},
	}
	if err := ctrl.SetControllerReference(&crd, secret, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Apply Secret
	// ... (simplified apply)
	existingSecret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKeyFromObject(secret), existingSecret)
	if err != nil && client.IgnoreNotFound(err) == nil {
		if err := r.Create(ctx, secret); err != nil {
			return ctrl.Result{}, err
		}
	} else if err == nil {
		existingSecret.Data = secret.Data
		if err := r.Update(ctx, existingSecret); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update Status
	if err := r.Status().Update(ctx, &crd); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frkrv1.FrkrClient{}).
		Complete(r)
}

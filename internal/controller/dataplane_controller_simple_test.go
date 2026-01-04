package controller

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

// Simple unit test example without ginkgo/gomega
func TestDataPlaneReconciler_Simple(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := frkrv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add scheme: %v", err)
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&frkrv1.FrkrDataPlane{}).
		Build()

	reconciler := &DataPlaneReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	t.Run("reconcile non-existent resource", func(t *testing.T) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "non-existent",
				Namespace: "default",
			},
		}

		ctx := context.Background()
		result, err := reconciler.Reconcile(ctx, req)
		if err != nil {
			t.Errorf("Reconcile() error = %v, want nil", err)
		}
		if result.Requeue {
			t.Errorf("Reconcile() should not requeue for non-existent resource")
		}
	})

	t.Run("reconcile existing resource", func(t *testing.T) {
		dataPlane := &frkrv1.FrkrDataPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-datapane",
				Namespace: "default",
			},
			Spec: frkrv1.FrkrDataPlaneSpec{
				PostgresConfig: frkrv1.DatabaseConfig{
					Host:        "localhost",
					Port:        5432,
					Database:    "testdb",
					User:        "testuser",
					PasswordRef: "postgres-secret",
				},
				BrokerConfig: frkrv1.MessageQueueConfig{
					Brokers: []string{"localhost:9092"},
				},
			},
		}

		if err := fakeClient.Create(context.Background(), dataPlane); err != nil {
			t.Fatalf("failed to create test resource: %v", err)
		}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "test-datapane",
				Namespace: "default",
			},
		}

		ctx := context.Background()
		result, err := reconciler.Reconcile(ctx, req)
		if err != nil {
			t.Errorf("Reconcile() error = %v, want nil", err)
		}
		if result.Requeue {
			t.Errorf("Reconcile() should not requeue")
		}

		// Verify status was updated
		updated := &frkrv1.FrkrDataPlane{}
		if err := fakeClient.Get(ctx, req.NamespacedName, updated); err != nil {
			t.Fatalf("failed to get updated resource: %v", err)
		}

		if updated.Status.Phase != "Active" {
			t.Errorf("expected phase Active, got %s", updated.Status.Phase)
		}
		if !updated.Status.PostgresConnected {
			t.Errorf("expected PostgresConnected to be true")
		}
		if !updated.Status.BrokerConnected {
			t.Errorf("expected BrokerConnected to be true")
		}
		if len(updated.Status.Warnings) > 0 {
			t.Errorf("expected no warnings, got %v", updated.Status.Warnings)
		}
	})
}

package controller

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

func TestDataPlaneReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DataPlane Controller Suite")
}

var _ = Describe("DataPlaneReconciler", func() {
	var (
		ctx        context.Context
		cancel     context.CancelFunc
		reconciler *DataPlaneReconciler
		fakeClient client.Client
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		scheme = runtime.NewScheme()
		_ = frkrv1.AddToScheme(scheme)

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&frkrv1.FrkrDataPlane{}).
			Build()

		reconciler = &DataPlaneReconciler{
			Client: fakeClient,
			Scheme: scheme,
		}
	})

	AfterEach(func() {
		cancel()
	})

	Describe("Reconcile", func() {
		Context("when DataPlane resource exists", func() {
			var dataPlane *frkrv1.FrkrDataPlane

			BeforeEach(func() {
				dataPlane = &frkrv1.FrkrDataPlane{
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
				Expect(fakeClient.Create(ctx, dataPlane)).To(Succeed())
			})

			It("should update status with connection information", func() {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-datapane",
						Namespace: "default",
					},
				}

				result, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				// Verify status was updated
				updated := &frkrv1.FrkrDataPlane{}
				Expect(fakeClient.Get(ctx, req.NamespacedName, updated)).To(Succeed())
				Expect(updated.Status.Phase).To(Equal("Active"))
				Expect(updated.Status.PostgresConnected).To(BeTrue())
				Expect(updated.Status.BrokerConnected).To(BeTrue())
			})

			It("should clear warnings when connections are healthy", func() {
				// First reconcile with warnings
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-datapane",
						Namespace: "default",
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify warnings are empty when connections are healthy
				updated := &frkrv1.FrkrDataPlane{}
				Expect(fakeClient.Get(ctx, req.NamespacedName, updated)).To(Succeed())
				Expect(updated.Status.Warnings).To(BeEmpty())
			})
		})

		Context("when DataPlane resource does not exist", func() {
			It("should not return an error", func() {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "non-existent",
						Namespace: "default",
					},
				}

				result, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
			})
		})
	})

	Describe("SetupWithManager", func() {
		It("should setup controller with manager", func() {
			// Skip this test in unit test environment (requires real kubeconfig)
			// For integration tests, use envtest to create a test manager
			Skip("requires envtest setup for integration testing")
		})
	})
})

// Table-driven test example
func TestDataPlaneReconciler_Reconcile(t *testing.T) {
	tests := []struct {
		name      string
		dataPlane *frkrv1.FrkrDataPlane
		wantErr   bool
		validate  func(*testing.T, *frkrv1.FrkrDataPlane)
	}{
		{
			name: "successful reconciliation",
			dataPlane: &frkrv1.FrkrDataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: frkrv1.FrkrDataPlaneSpec{
					PostgresConfig: frkrv1.DatabaseConfig{
						Host:        "localhost",
						Port:        5432,
						Database:    "testdb",
						User:        "testuser",
						PasswordRef: "secret",
					},
					BrokerConfig: frkrv1.MessageQueueConfig{
						Brokers: []string{"localhost:9092"},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, dp *frkrv1.FrkrDataPlane) {
				if dp.Status.Phase != "Active" {
					t.Errorf("expected phase Active, got %s", dp.Status.Phase)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = frkrv1.AddToScheme(scheme)

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&frkrv1.FrkrDataPlane{}).
				WithObjects(tt.dataPlane).
				Build()

			reconciler := &DataPlaneReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.dataPlane.Name,
					Namespace: tt.dataPlane.Namespace,
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := reconciler.Reconcile(ctx, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				updated := &frkrv1.FrkrDataPlane{}
				if err := fakeClient.Get(ctx, req.NamespacedName, updated); err != nil {
					t.Fatalf("failed to get updated resource: %v", err)
				}
				tt.validate(t, updated)
			}
		})
	}
}

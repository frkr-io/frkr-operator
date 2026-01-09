package controller

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
)

var _ = Describe("UserReconciler", func() {
	var (
		ctx        context.Context
		cancel     context.CancelFunc
		reconciler *UserReconciler
		fakeClient client.Client
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		scheme = runtime.NewScheme()
		_ = frkrv1.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&frkrv1.FrkrUser{}).
			Build()

		reconciler = &UserReconciler{
			Client: fakeClient,
			Scheme: scheme,
		}
	})

	AfterEach(func() {
		cancel()
	})

	Describe("Reconcile", func() {
		Context("when creating a new user", func() {
			var user *frkrv1.FrkrUser

			BeforeEach(func() {
				user = &frkrv1.FrkrUser{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-user",
						Namespace: "default",
					},
					Spec: frkrv1.FrkrUserSpec{
						Username: "testuser",
						TenantID: "tenant-1",
					},
				}
				Expect(fakeClient.Create(ctx, user)).To(Succeed())
			})

			It("should generate a password when not provided", func() {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-user",
						Namespace: "default",
					},
				}

				result, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				// Verify password was generated
				updated := &frkrv1.FrkrUser{}
				Expect(fakeClient.Get(ctx, req.NamespacedName, updated)).To(Succeed())
				Expect(updated.Status.Password).NotTo(BeEmpty())
				Expect(updated.Status.PasswordGenerated).To(BeTrue())
				Expect(updated.Status.Phase).To(Equal("Active"))
			})

			It("should create a secret with credentials", func() {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-user",
						Namespace: "default",
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret was created
				secret := &corev1.Secret{}
				secretName := types.NamespacedName{
					Name:      "frkr-user-testuser",
					Namespace: "default",
				}
				Expect(fakeClient.Get(ctx, secretName, secret)).To(Succeed())
				Expect(secret.Data["username"]).To(Equal([]byte("testuser")))
				Expect(secret.Data["password"]).NotTo(BeEmpty())
			})

			It("should use provided password if specified", func() {
				user.Spec.Password = "provided-password"
				Expect(fakeClient.Update(ctx, user)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-user",
						Namespace: "default",
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				updated := &frkrv1.FrkrUser{}
				Expect(fakeClient.Get(ctx, req.NamespacedName, updated)).To(Succeed())
				Expect(updated.Status.Password).To(Equal("provided-password"))
				Expect(updated.Status.PasswordGenerated).To(BeFalse())
			})
		})

		Context("when updating an existing user", func() {
			var user *frkrv1.FrkrUser
			var secret *corev1.Secret

			BeforeEach(func() {
				user = &frkrv1.FrkrUser{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-user",
						Namespace: "default",
					},
					Spec: frkrv1.FrkrUserSpec{
						Username: "testuser",
						TenantID: "tenant-1",
						Password: "old-password",
					},
				}
				Expect(fakeClient.Create(ctx, user)).To(Succeed())

				// Create existing secret
				secret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "frkr-user-testuser",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"username": []byte("testuser"),
						"password": []byte("old-password"),
					},
				}
				Expect(fakeClient.Create(ctx, secret)).To(Succeed())
			})

			It("should update the secret when password changes", func() {
				user.Spec.Password = "new-password"
				Expect(fakeClient.Update(ctx, user)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-user",
						Namespace: "default",
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret was updated
				updatedSecret := &corev1.Secret{}
				secretName := types.NamespacedName{
					Name:      "frkr-user-testuser",
					Namespace: "default",
				}
				Expect(fakeClient.Get(ctx, secretName, updatedSecret)).To(Succeed())
				Expect(updatedSecret.Data["password"]).To(Equal([]byte("new-password")))
			})
		})

		Context("when user does not exist", func() {
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
})

// Integration test example using envtest
func TestUserReconciler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Skip("requires envtest setup - see TESTING.md for details")

	// Example structure:
	// 1. Setup envtest environment
	// 2. Create manager and controller
	// 3. Create test resources
	// 4. Run reconciliation
	// 5. Verify results
	// 6. Cleanup
}

// Benchmark test
func BenchmarkUserReconciler_Reconcile(b *testing.B) {
	scheme := runtime.NewScheme()
	_ = frkrv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	user := &frkrv1.FrkrUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bench-user",
			Namespace: "default",
		},
		Spec: frkrv1.FrkrUserSpec{
			Username: "benchuser",
			TenantID: "tenant-1",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&frkrv1.FrkrUser{}).
		WithObjects(user).
		Build()

	reconciler := &UserReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "bench-user",
			Namespace: "default",
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reconciler.Reconcile(ctx, req)
	}
}

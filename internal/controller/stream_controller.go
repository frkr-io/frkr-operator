package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	frkrv1 "github.com/frkr-io/frkr-operator/api/v1"
	"github.com/frkr-io/frkr-operator/internal/infra"
)

// StreamReconciler reconciles a FrkrStream object
type StreamReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	DB         *infra.DB
	KafkaAdmin *infra.KafkaAdmin
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

	logger.Info("reconciling stream", "name", stream.Spec.Name, "tenantId", stream.Spec.TenantID)

	// Check if infrastructure is available
	if r.DB == nil {
		logger.Info("database connection not available, requeueing")
		r.updateStatus(ctx, &stream, "Pending", metav1.ConditionFalse, "InfrastructureNotReady", "Waiting for database connection")
		return ctrl.Result{RequeueAfter: 5}, nil
	}

	// Step 1: Ensure tenant exists
	tenantID, err := r.DB.EnsureTenant(stream.Spec.TenantID)
	if err != nil {
		logger.Error(err, "failed to ensure tenant")
		r.updateStatus(ctx, &stream, "Error", metav1.ConditionFalse, "TenantError", err.Error())
		return ctrl.Result{RequeueAfter: 30}, nil
	}

	// Step 2: Create stream record in database
	retentionDays := stream.Spec.RetentionDays
	if retentionDays == 0 {
		retentionDays = 7 // default
	}

	streamID, topic, err := r.DB.CreateStream(tenantID, stream.Spec.Name, stream.Spec.Description, retentionDays)
	if err != nil {
		logger.Error(err, "failed to create stream in database")
		r.updateStatus(ctx, &stream, "Error", metav1.ConditionFalse, "DatabaseError", err.Error())
		return ctrl.Result{RequeueAfter: 30}, nil
	}

	// Step 3: Create Kafka topic
	if r.KafkaAdmin != nil {
		if err := r.KafkaAdmin.CreateTopic(topic, 1, 1); err != nil {
			logger.Error(err, "failed to create Kafka topic", "topic", topic)
			r.updateStatus(ctx, &stream, "Error", metav1.ConditionFalse, "KafkaError", err.Error())
			return ctrl.Result{RequeueAfter: 30}, nil
		}
		logger.Info("kafka topic created/verified", "topic", topic)
	}

	// Step 4: Update status with success
	stream.Status.Phase = "Ready"
	stream.Status.StreamID = streamID
	stream.Status.Topic = topic
	meta.SetStatusCondition(&stream.Status.Conditions, metav1.Condition{
		Type:               "StreamCreated",
		Status:             metav1.ConditionTrue,
		Reason:             "Success",
		Message:            fmt.Sprintf("Stream created with topic: %s", topic),
		LastTransitionTime: metav1.Now(),
	})

	if err := r.Status().Update(ctx, &stream); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("stream reconciled successfully", "name", stream.Spec.Name, "topic", topic)
	return ctrl.Result{}, nil
}

func (r *StreamReconciler) updateStatus(ctx context.Context, stream *frkrv1.FrkrStream, phase string, conditionStatus metav1.ConditionStatus, reason, message string) {
	stream.Status.Phase = phase
	meta.SetStatusCondition(&stream.Status.Conditions, metav1.Condition{
		Type:               "StreamCreated",
		Status:             conditionStatus,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
	_ = r.Status().Update(ctx, stream)
}

// SetupWithManager sets up the controller with the Manager
func (r *StreamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frkrv1.FrkrStream{}).
		Complete(r)
}

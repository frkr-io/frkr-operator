package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FrkrStreamSpec defines the desired state of FrkrStream
type FrkrStreamSpec struct {
	// TenantID is the tenant/organization ID this stream belongs to
	TenantID string `json:"tenantId"`

	// Name is the stream name
	Name string `json:"name"`

	// Description is an optional description
	// +optional
	Description string `json:"description,omitempty"`

	// RetentionDays is the retention period in days
	// +optional
	RetentionDays int `json:"retentionDays,omitempty"`
}

// FrkrStreamStatus defines the observed state of FrkrStream
type FrkrStreamStatus struct {
	// Phase indicates the current phase
	// +optional
	Phase string `json:"phase,omitempty"`

	// StreamID is the generated stream ID
	// +optional
	StreamID string `json:"streamId,omitempty"`

	// Topic is the generated Kafka-compatible topic name
	// +optional
	Topic string `json:"topic,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Name",type="string",JSONPath=".spec.name"
//+kubebuilder:printcolumn:name="Tenant",type="string",JSONPath=".spec.tenantId"
//+kubebuilder:printcolumn:name="Topic",type="string",JSONPath=".status.topic"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// FrkrStream is the Schema for the frkrstreams API
type FrkrStream struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrkrStreamSpec   `json:"spec,omitempty"`
	Status FrkrStreamStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrkrStreamList contains a list of FrkrStream
type FrkrStreamList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrkrStream `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrkrStream{}, &FrkrStreamList{})
}

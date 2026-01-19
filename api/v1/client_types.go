package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FrkrClientSpec defines the desired state of FrkrClient
type FrkrClientSpec struct {
	// TenantID is the UUID of the tenant
	TenantID string `json:"tenantId"`

	// ClientID is the desired client ID string
	ClientID string `json:"clientId"`

	// Optional: StreamID to scope this client to
	// +optional
	StreamID string `json:"streamId,omitempty"`

	// Optional: Secret is the client credentials secret
	// If empty, one will be generated
	// +optional
	Secret string `json:"secret,omitempty"`
}

// FrkrClientStatus defines the observed state of FrkrClient
type FrkrClientStatus struct {
	// ID is the database UUID of the client credential
	ID string `json:"id,omitempty"`

	// Phase represents the current lifecycle state
	Phase string `json:"phase,omitempty"`

	// SecretGenerated indicates if the secret was auto-generated
	SecretGenerated bool `json:"secretGenerated,omitempty"`

	// Conditions store the status conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ID",type=string,JSONPath=`.status.id`
//+kubebuilder:printcolumn:name="ClientID",type=string,JSONPath=`.spec.clientId`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// FrkrClient is the Schema for the frkrclients API
type FrkrClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrkrClientSpec   `json:"spec,omitempty"`
	Status FrkrClientStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrkrClientList contains a list of FrkrClient
type FrkrClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrkrClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrkrClient{}, &FrkrClientList{})
}

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FrkrTenantSpec defines the desired state of FrkrTenant
type FrkrTenantSpec struct {
	// Name is the display name of the tenant (if different from metadata.name)
	// +optional
	Name string `json:"name,omitempty"`

	// Plan is the subscription plan (default: free)
	// +optional
	Plan string `json:"plan,omitempty"`
}

// FrkrTenantStatus defines the observed state of FrkrTenant
type FrkrTenantStatus struct {
	// ID is the database UUID of the tenant
	ID string `json:"id,omitempty"`

	// Phase represents the current lifecycle state
	Phase string `json:"phase,omitempty"`

	// Conditions store the status conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ID",type=string,JSONPath=`.status.id`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// FrkrTenant is the Schema for the frkrtenants API
type FrkrTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrkrTenantSpec   `json:"spec,omitempty"`
	Status FrkrTenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrkrTenantList contains a list of FrkrTenant
type FrkrTenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrkrTenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrkrTenant{}, &FrkrTenantList{})
}

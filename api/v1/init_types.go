package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FrkrInitSpec defines the desired state of FrkrInit
type FrkrInitSpec struct {
	// MigrationsPath is the path to migration files
	// +optional
	MigrationsPath string `json:"migrationsPath,omitempty"`

	// DatabaseURL is the database connection URL
	// If not provided, will use FrkrDataPlane configuration
	// +optional
	DatabaseURL string `json:"databaseUrl,omitempty"`

	// Gateways is a list of Gateway Deployments to verify after migration
	// +optional
	Gateways []string `json:"gateways,omitempty"`
}

// FrkrInitStatus defines the observed state of FrkrInit
type FrkrInitStatus struct {
	// Phase indicates the current phase
	// +optional
	Phase string `json:"phase,omitempty"`

	// Version is the current migration version
	// +optional
	Version uint `json:"version,omitempty"`

	// Dirty indicates if the database is in a dirty state
	// +optional
	Dirty bool `json:"dirty,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Version",type="integer",JSONPath=".status.version"
//+kubebuilder:printcolumn:name="Dirty",type="boolean",JSONPath=".status.dirty"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// FrkrInit is the Schema for the frkrinits API
// This CRD replaces frkr-init-core-stack
type FrkrInit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrkrInitSpec   `json:"spec,omitempty"`
	Status FrkrInitStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrkrInitList contains a list of FrkrInit
type FrkrInitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrkrInit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrkrInit{}, &FrkrInitList{})
}

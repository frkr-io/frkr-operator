package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FrkrDataPlaneSpec defines the desired state of FrkrDataPlane
type FrkrDataPlaneSpec struct {
	// PostgresConfig is the PostgreSQL/CockroachDB configuration
	PostgresConfig DatabaseConfig `json:"postgresConfig"`
	
	// RedpandaConfig is the Redpanda/Kafka configuration
	RedpandaConfig MessageQueueConfig `json:"redpandaConfig"`
}

// DatabaseConfig defines database connection configuration
type DatabaseConfig struct {
	// Host is the database host
	Host string `json:"host"`
	
	// Port is the database port
	// +optional
	Port int `json:"port,omitempty"`
	
	// Database is the database name
	Database string `json:"database"`
	
	// User is the database user
	User string `json:"user"`
	
	// PasswordRef is a reference to a secret containing the password
	PasswordRef string `json:"passwordRef"`
	
	// SSLMode is the SSL mode (require, disable, etc.)
	// +optional
	SSLMode string `json:"sslMode,omitempty"`
}

// MessageQueueConfig defines message queue configuration
type MessageQueueConfig struct {
	// Brokers is a list of broker addresses
	Brokers []string `json:"brokers"`
	
	// TLSEnabled indicates if TLS is enabled
	// +optional
	TLSEnabled bool `json:"tlsEnabled,omitempty"`
	
	// TLSConfigRef is a reference to a secret containing TLS certificates
	// +optional
	TLSConfigRef string `json:"tlsConfigRef,omitempty"`
}

// FrkrDataPlaneStatus defines the observed state of FrkrDataPlane
type FrkrDataPlaneStatus struct {
	// Phase indicates the current phase
	// +optional
	Phase string `json:"phase,omitempty"`
	
	// PostgresConnected indicates if Postgres connection is healthy
	// +optional
	PostgresConnected bool `json:"postgresConnected,omitempty"`
	
	// RedpandaConnected indicates if Redpanda connection is healthy
	// +optional
	RedpandaConnected bool `json:"redpandaConnected,omitempty"`
	
	// Warnings contains any connectivity warnings
	// +optional
	Warnings []string `json:"warnings,omitempty"`
	
	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Postgres",type="string",JSONPath=".status.postgresConnected"
//+kubebuilder:printcolumn:name="Redpanda",type="string",JSONPath=".status.redpandaConnected"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// FrkrDataPlane is the Schema for the frkrdatapanes API
type FrkrDataPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrkrDataPlaneSpec   `json:"spec,omitempty"`
	Status FrkrDataPlaneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrkrDataPlaneList contains a list of FrkrDataPlane
type FrkrDataPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrkrDataPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrkrDataPlane{}, &FrkrDataPlaneList{})
}


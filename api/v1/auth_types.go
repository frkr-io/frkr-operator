package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AuthType defines the authentication type
// +kubebuilder:validation:Enum=basic;oidc
type AuthType string

const (
	AuthTypeBasic AuthType = "basic"
	AuthTypeOIDC  AuthType = "oidc"
)

// FrkrAuthConfigSpec defines the desired state of FrkrAuthConfig
type FrkrAuthConfigSpec struct {
	// Type is the authentication type (basic or oidc)
	Type AuthType `json:"type"`

	// OIDCConfig is the OIDC configuration (required if type is oidc)
	// +optional
	OIDCConfig *OIDCConfig `json:"oidcConfig,omitempty"`
}

// OIDCConfig defines OIDC provider configuration
type OIDCConfig struct {
	// IssuerURL is the OIDC issuer URL
	IssuerURL string `json:"issuerUrl"`

	// ClientID is the OIDC client ID
	ClientID string `json:"clientId"`

	// ClientSecret is the OIDC client secret (stored in secret)
	ClientSecretRef string `json:"clientSecretRef"`

	// Scopes are the OIDC scopes to request
	// +optional
	Scopes []string `json:"scopes,omitempty"`
}

// FrkrAuthConfigStatus defines the observed state of FrkrAuthConfig
type FrkrAuthConfigStatus struct {
	// Phase indicates the current phase of the auth configuration
	// +optional
	Phase string `json:"phase,omitempty"`

	// PreviousType is the previous auth type (used for cleanup)
	// +optional
	PreviousType AuthType `json:"previousType,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// FrkrAuthConfig is the Schema for the frkrauthconfigs API
type FrkrAuthConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrkrAuthConfigSpec   `json:"spec,omitempty"`
	Status FrkrAuthConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrkrAuthConfigList contains a list of FrkrAuthConfig
type FrkrAuthConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrkrAuthConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrkrAuthConfig{}, &FrkrAuthConfigList{})
}

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FrkrUserSpec defines the desired state of FrkrUser
type FrkrUserSpec struct {
	// Username is the user's username
	Username string `json:"username"`
	
	// Password is optional. If not provided, a random password will be generated
	// +optional
	Password string `json:"password,omitempty"`
	
	// TenantID is the tenant/organization ID this user belongs to
	TenantID string `json:"tenantId"`
	
	// Roles are the user's roles
	// +optional
	Roles []string `json:"roles,omitempty"`
}

// FrkrUserStatus defines the observed state of FrkrUser
type FrkrUserStatus struct {
	// Phase indicates the current phase of the user
	// +optional
	Phase string `json:"phase,omitempty"`
	
	// Password is the generated or provided password (one-time display only)
	// This field is populated only once when the user is created
	// +optional
	Password string `json:"password,omitempty"`
	
	// PasswordGenerated indicates if the password was auto-generated
	// +optional
	PasswordGenerated bool `json:"passwordGenerated,omitempty"`
	
	// LastPasswordReset is the timestamp of the last password reset
	// +optional
	LastPasswordReset *metav1.Time `json:"lastPasswordReset,omitempty"`
	
	// Conditions represent the latest available observations of the user's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Username",type="string",JSONPath=".spec.username"
//+kubebuilder:printcolumn:name="Tenant",type="string",JSONPath=".spec.tenantId"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// FrkrUser is the Schema for the frkrusers API
type FrkrUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrkrUserSpec   `json:"spec,omitempty"`
	Status FrkrUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrkrUserList contains a list of FrkrUser
type FrkrUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrkrUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrkrUser{}, &FrkrUserList{})
}


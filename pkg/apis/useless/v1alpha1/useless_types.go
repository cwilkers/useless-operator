package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UselessSpec defines the desired state of Useless
// +k8s:openapi-gen=true
type UselessSpec struct {
	// Size is the desired number of pods
        Size int32 `json:"size"`
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// UselessStatus defines the observed state of Useless
// +k8s:openapi-gen=true
type UselessStatus struct {
	// Nodes are the names of pods belonging to the useless machine
        Nodes []string `json:"nodes"`
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Useless is the Schema for the uselesses API
// +k8s:openapi-gen=true
type Useless struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UselessSpec   `json:"spec,omitempty"`
	Status UselessStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UselessList contains a list of Useless
type UselessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Useless `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Useless{}, &UselessList{})
}

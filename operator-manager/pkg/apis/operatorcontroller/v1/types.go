package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Operator is a specification for a Operator resource
// +k8s:openapi-gen=true
type Operator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatorSpec   `json:"spec"`
	Status OperatorStatus `json:"status"`
}

// OperatorSpec is the spec for a OperatorSpec resource
// +k8s:openapi-gen=true
type OperatorSpec struct {
	Name string `json:"name"`
	ChartURL string `json:"chartURL"`
	Values map[string]interface{} `json:"values"`
}

// OperatorStatus is the status for a Operator resource
// +k8s:openapi-gen=true
type OperatorStatus struct {
	CustomResourceDefinitions []string `json:"customResourceDefinitions"`
	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// OperatorList is a list of Operator resources
type OperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Operator `json:"items"`
}

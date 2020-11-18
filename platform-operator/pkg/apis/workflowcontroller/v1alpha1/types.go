package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PlatformWorkflow is specification for a PlatformWorkflow resource
// +k8s:openapi-gen=true
type PlatformWorkflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlatformWorkflowSpec   `json:"spec"`
	Status PlatformWorkflowStatus `json:"status"`
}

// PlatformWorkflowSpec is the spec for a PlatformWorkflow resource
// +k8s:openapi-gen=true
type PlatformWorkflowSpec struct {
	// LabelSelector that selects resources of that label
	LabelSelector map[string]string `json:"labelSelector,omitempty"` 
	// List of stack elements that forms this Platform Workflow
	StackElements []StackElements `json:"stackElements"`
	// Name of CRD to register
	CustomAPI []CustomAPI `json:"customAPI"`
}

type StackElements struct {
	// Name of Kubernetes Kind - could be Built-in Kind or Custom Kind
	Kind string `json:"kind"`
	// Name of the Resource
	Name string `json:"name"`
	// Namespace of the Resource
	// If not specified then 'default' Namespace is assumed
	// +optional
	Namespace string `json:"namespace"`
	// +optional
	DependsOn []DependsOn `json:"dependsOn,omitempty"`
}

type DependsOn struct {
	// Name of the Resource that the Resource in which dependsOn is included actually dependsOn
	Name string `json:"name"`
}

type CustomAPI struct {
	// Kind of the Custom API
	Kind string `json:"kind"`
	// Version of the API Custom API
	Version string `json:"version"`
	// Group of the Custom API
	Group string `json:"group"`
	// Plural name for the Custom API
	Plural string `json:"plural"`
	// Helm chart URL
	ChartURL string `json:"chartURL"`
	// Chart name
	ChartName string `json:"chartName"`
	// Values
	//Values []Values `json:"values,omitempty"`
}

type Values struct {
	// Name - as used in values.yaml
	Name string `json:"name"`
	// Value - as specified in values.yaml
	Value string `json:"value"`
}

// PlatformWorkflowStatus is the status for a PlatformWorkflow resource.
// +k8s:openapi-gen=true
type PlatformWorkflowStatus struct {
	Status             string   `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PlatformWorkflowList is a list of PlatformWorkflow resources
type PlatformWorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PlatformWorkflow `json:"items"`
}
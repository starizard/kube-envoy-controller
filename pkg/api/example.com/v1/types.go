package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=envoys

type Envoy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   EnvoySpec   `json:"spec"`
	Status EnvoyStatus `json:"status,omitempty"`
}

type EnvoySpec struct {
	Name          string `json:"name"`
	ConfigMapName string `json:"configMapName"`
	Replicas      *int32 `json:"replicas"`
}

type EnvoyStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EnvoyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Envoy `json:"items"`
}

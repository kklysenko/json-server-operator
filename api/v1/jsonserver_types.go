/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JsonServerSpec defines the desired state of JsonServer
type JsonServerSpec struct {
	// Replicas is the number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// JsonConfig is the JSON configuration for the json-server
	JsonConfig string `json:"jsonConfig,omitempty"`
}

type JsonServerState string

const (
	// SyncedJsonServerState indicating that the object was synced successfully
	SyncedJsonServerState JsonServerState = "Synced"

	// ErrorJsonServerState indicating that the object failed to sync
	ErrorJsonServerState JsonServerState = "Error"
)

// JsonServerStatus defines the observed state of JsonServer
type JsonServerStatus struct {
	// State indicates if the object was synced successfully
	// +optional
	State JsonServerState `json:"state,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// Replicas is the total number of non-terminated pods targeted by this deployment (their labels match the selector).
	// +optional
	Replicas int32 `json:"replicas,omitempty" protobuf:"varint,2,opt,name=replicas"`

	// Selector that identifies the pods that are receiving active traffic
	// +optional
	Selector string `json:"selector,omitempty"`
}

func NewJsonServerStatus(err error, deployment *appsv1.Deployment) *JsonServerStatus {
	state := SyncedJsonServerState
	message := "Synced successfully!"

	if err != nil {
		state = ErrorJsonServerState
		message = err.Error()
	}

	return &JsonServerStatus{
		State:    state,
		Message:  message,
		Replicas: deployment.Status.Replicas,
		Selector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	}
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:deepcopy-gen:true
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// JsonServer is the Schema for the jsonservers API
type JsonServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JsonServerSpec   `json:"spec,omitempty"`
	Status JsonServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// JsonServerList contains a list of JsonServer
type JsonServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JsonServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JsonServer{}, &JsonServerList{})
}

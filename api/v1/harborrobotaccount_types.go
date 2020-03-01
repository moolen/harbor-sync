/*
Copyright 2019 The Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HarborRobotAccountSpec defines the desired state of HarborRobotAccount
type HarborRobotAccountSpec struct {
	Credential RobotAccountCredential `json:"credential"`
}

// HarborRobotAccountStatus defines the observed state of HarborRobotAccount
type HarborRobotAccountStatus struct {
	LastSync int64 `json:"last_sync"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// HarborRobotAccount is the Schema for the harborrobotaccounts API
type HarborRobotAccount struct {
	// we use a specific label that allows us to search t
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborRobotAccountSpec   `json:"spec,omitempty"`
	Status HarborRobotAccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HarborRobotAccountList contains a list of HarborRobotAccount
type HarborRobotAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborRobotAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HarborRobotAccount{}, &HarborRobotAccountList{})
}

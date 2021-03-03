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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HarborRobotAccountSpec defines the desired state of HarborRobotAccount
type HarborRobotAccountSpec struct {
	Credential RobotAccountCredential `json:"credential"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// HarborRobotAccount is the Schema for the harborrobotaccounts API
type HarborRobotAccount struct {
	// we use a specific label that allows us to search t
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborRobotAccountSpec   `json:"spec,omitempty"`
	Status HarborRobotAccountStatus `json:"status,omitempty"`
}

// HarborRobotAccountStatus defines the observed state of HarborRobotAccount
type HarborRobotAccountStatus struct {
	// +nullable
	// refreshTime is the time and date the robot account was fetched and
	// the target secret updated
	RefreshTime metav1.Time `json:"refreshTime,omitempty"`

	// +optional
	Conditions []RobotAccountStatusCondition `json:"conditions,omitempty"`
}

type HarborRobotAccountConditionType string

const (
	RobotAccountReady HarborRobotAccountConditionType = "Ready"
)

type RobotAccountStatusCondition struct {
	Type   HarborRobotAccountConditionType `json:"type"`
	Status corev1.ConditionStatus          `json:"status"`

	// +optional
	Reason string `json:"reason,omitempty"`

	// +optional
	Message string `json:"message,omitempty"`

	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
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

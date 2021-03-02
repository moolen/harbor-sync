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

// HarborSyncSpec defines the desired state
// how should harbor projects map to secrets in namespaces
type HarborSyncSpec struct {

	// Specifies how to do matching on a harbor project.
	// Valid values are:
	// - "Regex" (default): interpret the project name as regular expression;
	Type ProjectMatchingType `json:"type"`

	// ProjectName specifies the project name
	ProjectName string `json:"name"`

	// PushAccess allows the robot account to push images, too. defaults to false.
	// As of now we can not tell what permissions a robot account has. The user
	// has to wait for the next rotation until the robot account has the new permissions.
	// Alternatively, you can re-create your HarborSync spec. This forces a rotation.
	PushAccess bool `json:"pushAccess"`

	// The RobotAccountSuffix specifies the suffix to use when creating a new robot account
	// +kubebuilder:validation:MinLength=4
	RobotAccountSuffix string `json:"robotAccountSuffix"`

	// The Mapping contains the mapping from project to a secret in a namespace
	Mapping []ProjectMapping `json:"mapping,omitempty"`

	// Webhook contains a list of endpoints which will be called
	// if the robot account changes (e..g automatic rotation, expired account, disabled...)
	// +optional
	Webhook []WebhookConfig `json:"webhook,omitempty"`
}

// ProjectMatchingType specifies the type of matching to be done.
// Only one of the following matching types may be specified.
// If none of the following types is specified, the default one
// is Regex.
// +kubebuilder:validation:Enum=Regex
type ProjectMatchingType string

const (
	// RegexMatching interprets the name field as regular expression
	// Capturing groups may be used in a ProjectMapping
	RegexMatching ProjectMatchingType = "Regex"
)

// ProjectMapping defines how projects are mapped to secrets in specific namespaces
type ProjectMapping struct {
	Namespace string      `json:"namespace"`
	Secret    string      `json:"secret"`
	Type      MappingType `json:"type"`
}

// MappingType specifies how to map the project into the namespace/secret
// Only one of the following matching types may be specified.
// If none of the following types is specified, the default one
// is Translate.
// +kubebuilder:validation:Enum=Translate;Match
type MappingType string

const (
	// TranslateMappingType interpolates the project expression into the namespace
	TranslateMappingType MappingType = "Translate"

	// MatchMappingType treats the Namespace as regular expression and injects secrets into
	// all matching namespaces
	MatchMappingType MappingType = "Match"
)

// WebhookConfig defines how to call a webhook
type WebhookConfig struct {
	// Endpoint is a url
	Endpoint string `json:"endpoint"`
}

// WebhookUpdatePayload contains the new credentials of a robot account
type WebhookUpdatePayload struct {
	Project     string                 `json:"project"`
	Credentials RobotAccountCredential `json:"credentials"`
}

// HarborSyncStatus defines the observed state of HarborSync
type HarborSyncStatus struct {
	ProjectList []string `json:"managedProjects,omitempty"`

	// +optional
	Conditions []HarborSyncStatusCondition `json:"conditions,omitempty"`
}

type HarborSyncConditionType string

const (
	HarborSyncReady HarborSyncConditionType = "Ready"
)

type HarborSyncStatusCondition struct {
	Type   HarborSyncConditionType `json:"type"`
	Status corev1.ConditionStatus  `json:"status"`

	// +optional
	Reason string `json:"reason,omitempty"`

	// +optional
	Message string `json:"message,omitempty"`

	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// RobotAccountCredential holds the robot account name & token to access the harbor API
// this is also part of the webhook API change here might impact downstream users
type RobotAccountCredential struct {
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	Token     string `json:"token"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// HarborSync is the Schema for the harborsyncs API
type HarborSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborSyncSpec   `json:"spec,omitempty"`
	Status HarborSyncStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HarborSyncList contains a list of HarborSync
type HarborSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HarborSync{}, &HarborSyncList{})
}

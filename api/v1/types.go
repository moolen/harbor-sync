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

// HarborSyncConfigSpec defines the desired state of HarborSyncConfig
type HarborSyncConfigSpec struct {

	// ProjectSelector specifies a list of projects to look up and synchronize
	ProjectSelector []ProjectSelector `json:"projectSelector"`
}

// ProjectSelector defines how to select harbor projects
type ProjectSelector struct {
	// Specifies how to do matching on a harbor project.
	// Valid values are:
	// - "Regex" (default): interpret the project name as regular expression;
	Type ProjectMatchingType `json:"type"`

	// ProjectName specifies the project name
	ProjectName string `json:"name"`

	// TODO:
	// AccessType: ReadOnly|ReadWrite

	// The RobotAccountSuffix specifies the suffix to use when creating a new robot account
	// +kubebuilder:validation:MinLength=4
	RobotAccountSuffix string `json:"robotAccountSuffix"`

	// The Mapping contains the mapping from project to a secret in a namespace
	Mapping []ProjectMapping `json:"mapping"`

	// Webhook is WIP
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
	// TODO: design doc
}

// HarborSyncConfigStatus defines the observed state of HarborSyncConfig
type HarborSyncConfigStatus struct {
	RobotCredentials map[string]RobotAccountCredentials `json:"credentials"`
}

// RobotAccountCredentials holds a list of credentials
type RobotAccountCredentials []RobotAccountCredential

// RobotAccountCredential holds the robot account name & token to access the harbor API
type RobotAccountCredential struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// HarborSyncConfig is the Schema for the harborsyncconfigs API
type HarborSyncConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborSyncConfigSpec   `json:"spec,omitempty"`
	Status HarborSyncConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HarborSyncConfigList contains a list of HarborSyncConfig
type HarborSyncConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborSyncConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HarborSyncConfig{}, &HarborSyncConfigList{})
}

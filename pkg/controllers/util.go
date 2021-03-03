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
package controllers

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
)

// UpdateProjectStatusLastReconciliation sets the time when the robot was synced
func UpdateProjectStatusLastReconciliation(status *crdv1.HarborSyncStatus, project harbor.Project) {
	for i, p := range status.ProjectList {
		if p.Name == project.Name {
			status.ProjectList[i].LastRobotReconciliation = metav1.Now()
			return
		}
	}
	status.ProjectList = append(status.ProjectList, crdv1.ProjectStatus{
		Name:                    project.Name,
		LastRobotReconciliation: metav1.Now(),
		ManagedNamespaces:       []string{},
	})
}

// NewSyncCondition returns a new condition
func NewSyncCondition(condType crdv1.HarborSyncConditionType, status v1.ConditionStatus, reason, message string) *crdv1.HarborSyncStatusCondition {
	return &crdv1.HarborSyncStatusCondition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetSyncCondition returns the condition with the provided type.
func GetSyncCondition(status crdv1.HarborSyncStatus, condType crdv1.HarborSyncConditionType) *crdv1.HarborSyncStatusCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// SetSyncCondition updates the HarborSync resource to include the provided
// condition.
func SetSyncCondition(status *crdv1.HarborSyncStatus, condition crdv1.HarborSyncStatusCondition) {
	currentCond := GetSyncCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return
	}
	// Do not update lastTransitionTime if the status of the condition doesn't change.
	if currentCond != nil && currentCond.Status == condition.Status {
		condition.LastTransitionTime = currentCond.LastTransitionTime
	}
	status.Conditions = append(filterOutCondition(status.Conditions, condition.Type), condition)
}

func filterOutCondition(conditions []crdv1.HarborSyncStatusCondition, condType crdv1.HarborSyncConditionType) []crdv1.HarborSyncStatusCondition {
	newConditions := make([]crdv1.HarborSyncStatusCondition, 0, len(conditions))
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}

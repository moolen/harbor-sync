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

package reconciler

import (
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
)

// UpdateProjectStatusNamespace updates the namespaces that are referenced via a sync config
func UpdateProjectStatusNamespace(status *crdv1.HarborSyncStatus, project harbor.Project, namespace string) {

	for i, p := range status.ProjectList {
		if p.Name == project.Name {
			if !contains(p.ManagedNamespaces, namespace) {
				status.ProjectList[i].ManagedNamespaces = append(status.ProjectList[i].ManagedNamespaces, namespace)
			}
			return
		}
	}
	status.ProjectList = append(status.ProjectList, crdv1.ProjectStatus{
		Name:              project.Name,
		ManagedNamespaces: []string{namespace},
	})
}

func contains(arr []string, el string) bool {
	for _, a := range arr {
		if a == el {
			return true
		}
	}
	return false
}

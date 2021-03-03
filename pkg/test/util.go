package test

import (
	crdv1 "github.com/moolen/harbor-sync/api/v1"
)

// CheckProjects checks if the status struct contains the expected project and namespaces
func CheckProjects(expected map[string][]string, status crdv1.HarborSyncStatus) bool {
	for eName, eNs := range expected {
		var found bool
		for _, p := range status.ProjectList {
			if p.Name == eName && len(p.ManagedNamespaces) > 0 {
				found = true
				for _, pNs := range p.ManagedNamespaces {
					if !Contains(eNs, pNs) {
						return false
					}
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Contains returns true if arr contains el
func Contains(arr []string, el string) bool {
	for _, a := range arr {
		if a == el {
			return true
		}
	}
	return false
}

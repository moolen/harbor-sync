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
	crdv1 "github.com/moolen/harbor-sync/api/v1"

	"context"
	"regexp"

	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/moolen/harbor-sync/pkg/harbor"
)

func (r *HarborSyncConfigReconciler) mapByMatching(mapping crdv1.ProjectMapping, matcher *regexp.Regexp, project harbor.Project, credential crdv1.RobotAccountCredential) {
	nsMatcher, err := regexp.Compile(mapping.Namespace)
	if err != nil {
		r.Log.Error(err, "invalid regex", "namespace", mapping.Namespace)
		return
	}
	// get all namespaces
	// match ns against mapping.Namespace regex
	var nsList v1.NamespaceList
	err = r.List(context.Background(), &nsList)
	if err != nil {
		r.Log.Error(err, "error listing namespaces")
		return
	}

	for _, ns := range nsList.Items {
		if nsMatcher.MatchString(ns.Name) {
			proposedSecret := matcher.ReplaceAllString(project.Name, mapping.Secret)
			secret := makeSecret(ns.Name, proposedSecret, r.Harbor.BaseURL(), credential)
			upsertSecret(r, r.Log, secret)
		}
	}
}

func (r *HarborSyncConfigReconciler) mapByTranslating(mapping crdv1.ProjectMapping, matcher *regexp.Regexp, project harbor.Project, credential crdv1.RobotAccountCredential) {
	// propse a namespace and secret name / ignore missing namespace
	var ns v1.Namespace
	proposedNamespace := matcher.ReplaceAllString(project.Name, mapping.Namespace)
	err := r.Get(context.Background(), types.NamespacedName{Name: proposedNamespace}, &ns)
	if apierrs.IsNotFound(err) {
		r.Log.V(1).Info("ignoring proposed namespace", "project_name", project.Name, "proposed_namespace", proposedNamespace)
		return
	} else if err != nil {
		r.Log.Error(err, "error fetching namespace", "proposed_namespace", proposedNamespace)
		return
	}

	// propose a secret name for this project
	proposedSecret := matcher.ReplaceAllString(project.Name, mapping.Secret)
	r.Log.V(2).Info("proposing secret", "project_name", project.Name, "proposed_namespace", proposedNamespace, "proposed_secret", proposedSecret)
	secret := makeSecret(proposedNamespace, proposedSecret, r.Harbor.BaseURL(), credential)
	upsertSecret(r, r.Log, secret)
}

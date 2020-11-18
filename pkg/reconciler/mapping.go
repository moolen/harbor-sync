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
	"fmt"
	"strings"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"context"
	"regexp"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/moolen/harbor-sync/pkg/harbor"
	"github.com/moolen/harbor-sync/pkg/util"
)

// MappingFunc implements a specific strategy for
// reconciling the cluster state
type MappingFunc func(
	client.Client,
	crdv1.ProjectMapping,
	crdv1.HarborSync,
	harbor.Project,
	crdv1.RobotAccountCredential,
	string) error

// MappingFuncForConfig returns a MappingFunc for the given mapping
// which can be used by the called to reconcile the desired state
func MappingFuncForConfig(mapping crdv1.ProjectMapping) (MappingFunc, error) {
	if mapping.Type == crdv1.TranslateMappingType {
		return mapByTranslating, nil
	} else if mapping.Type == crdv1.MatchMappingType {
		return mapByMatching, nil
	}
	return nil, fmt.Errorf("invalid mapping type: %s", mapping.Type)
}

func mapByMatching(
	cl client.Client,
	mapping crdv1.ProjectMapping,
	syncConfig crdv1.HarborSync,
	project harbor.Project,
	credential crdv1.RobotAccountCredential,
	harborURL string,
) error {
	nsMatcher, err := regexp.Compile(mapping.Namespace)
	if err != nil {
		return fmt.Errorf("invalid regex: %s", err.Error())
	}
	// get all namespaces
	// match ns against mapping.Namespace regex
	var nsList v1.NamespaceList
	err = cl.List(context.Background(), &nsList)
	if err != nil {
		return fmt.Errorf("error listingnamespaces: %s", err.Error())
	}

	matcher, err := regexp.Compile(syncConfig.Spec.ProjectName)
	if err != nil {
		return fmt.Errorf("error compiling regex: %s", err.Error())
	}

	var errs []string
	for _, ns := range nsList.Items {
		if nsMatcher.MatchString(ns.Name) {
			proposedSecret := matcher.ReplaceAllString(project.Name, mapping.Secret)
			secret := util.MakeSecret(ns.Name, proposedSecret, harborURL, credential)
			err = util.UpsertSecret(cl, secret)
			if err != nil {
				errs = append(errs, err.Error())
				continue
			}
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("error upserting secrets: %s", strings.Join(errs, " | "))
}

func mapByTranslating(
	cl client.Client,
	mapping crdv1.ProjectMapping,
	syncConfig crdv1.HarborSync,
	project harbor.Project,
	credential crdv1.RobotAccountCredential,
	harborURL string,
) error {
	matcher, err := regexp.Compile(syncConfig.Spec.ProjectName)
	if err != nil {
		return fmt.Errorf("error compiling regex: %s", err.Error())
	}
	// propse a namespace and secret name / ignore missing namespace
	var ns v1.Namespace
	proposedNamespace := matcher.ReplaceAllString(project.Name, mapping.Namespace)
	err = cl.Get(context.Background(), types.NamespacedName{Name: proposedNamespace}, &ns)
	if apierrs.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error fetching namespace %s: %s", proposedNamespace, err.Error())
	}

	// propose a secret name for this project
	proposedSecret := matcher.ReplaceAllString(project.Name, mapping.Secret)
	secret := util.MakeSecret(proposedNamespace, proposedSecret, harborURL, credential)
	return util.UpsertSecret(cl, secret)
}

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
	"context"
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
)

// HarborSyncConfigReconciler reconciles a HarborSyncConfig object
type HarborSyncConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Harbor harbor.API
}

// +kubebuilder:rbac:groups=crd.harborsync.k8s.io,resources=harborsyncconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.harborsync.k8s.io,resources=harborsyncconfigs/status,verbs=get;update;patch

func (r *HarborSyncConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("harborsyncconfig", req.NamespacedName)

	var syncConfig crdv1.HarborSyncConfig
	if err := r.Get(ctx, req.NamespacedName, &syncConfig); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("ignoring object delete")
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch sync config")
		return ctrl.Result{}, err
	}

	// for each project in projectSelector
	//
	// - find projects that match the selector
	// - if match:
	//   - reconcile robot account
	//   - populate secret in specified namespace
	for _, selector := range syncConfig.Spec.ProjectSelector {
		log.Info("iterating", "type", selector.Type, "project_name", selector.ProjectName)
		var err error
		var matchingProjects []harbor.Project
		var matcher *regexp.Regexp

		allProjects, err := r.Harbor.ListProjects()
		if err != nil {
			log.Error(err, "could not list harbor projects")
			continue
		}
		if selector.Type != crdv1.RegexMatching {
			log.Error(fmt.Errorf("invalid selector type: %s", selector.Type), "selector type must be regex")
			continue
		}
		matcher, err = regexp.Compile(selector.ProjectName)
		if err != nil {
			log.Error(err, "error compiling regex", "selector_project_name", selector.ProjectName)
			continue
		}
		for _, project := range allProjects {
			if matcher.MatchString(project.Name) {
				log.Info("project match", "type", selector.Type, "project_name", project.Name)
				matchingProjects = append(matchingProjects, project)
			}
		}

		// check if projects have a specific robot account
		// create it if not
		for _, project := range matchingProjects {
			skip, projectCredential := reconcileRobotAccounts(r.Harbor, log, &syncConfig, project, selector)
			if skip {
				continue
			}

			// reconcile secrets in namespaces
			for _, mapping := range selector.Mapping {
				if mapping.Type == crdv1.TranslateMappingType {
					r.mapByTranslating(mapping, matcher, project, *projectCredential)
				} else if mapping.Type == crdv1.MatchMappingType {
					r.mapByMatching(mapping, matcher, project, *projectCredential)
				} else {
					// not implemented
					log.Error(fmt.Errorf("invalid mapping type: %s", mapping.Type), "unsupported mapping")
				}
			}
		}
	}

	err := r.Update(context.Background(), &syncConfig)
	if err != nil {
		log.Error(err, "could not update syncConfig status field", "sync_config_name", syncConfig.Name)
	}
	return ctrl.Result{}, nil
}

func (r *HarborSyncConfigReconciler) SetupWithManager(mgr ctrl.Manager, input <-chan event.GenericEvent) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.HarborSyncConfig{}).
		Watches(&source.Channel{Source: input}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

func makeSecret(namespace, name string, baseURL string, credentials crdv1.RobotAccountCredential) v1.Secret {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", credentials.Name, credentials.Token)))
	configJSON := fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","auth":"%s"}}}`, baseURL, credentials.Name, credentials.Token, auth)
	return v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Type: v1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			v1.DockerConfigJsonKey: []byte(configJSON),
		},
	}
}

func (r *HarborSyncConfigReconciler) upsertSecret(secret v1.Secret) {
	err := r.Create(context.Background(), &secret)
	if apierrs.IsAlreadyExists(err) {
		err = r.Update(context.TODO(), &secret)
		if err != nil {
			r.Log.Error(err, "could not update secret", "proposed_namespace", secret.ObjectMeta.Namespace, "proposed_secret", secret.ObjectMeta.Name)
			return
		}
		r.Log.Info("updated secret", "proposed_namespace", secret.ObjectMeta.Namespace, "proposed_secret", secret.ObjectMeta.Name)
		return
	}
	if err != nil {
		r.Log.Error(err, "could not create secret", "proposed_namespace", secret.ObjectMeta.Namespace, "proposed_secret", secret.ObjectMeta.Name)
		return
	}
	r.Log.Info("created secret", "proposed_namespace", secret.ObjectMeta.Namespace, "proposed_secret", secret.ObjectMeta.Name)
	return
}

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

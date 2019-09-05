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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/go-logr/logr"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
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

// Reconcile reconciles the desired state in the cluster
func (r *HarborSyncConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("harborsyncconfig", req.NamespacedName)

	var syncConfig crdv1.HarborSync
	if err := r.Get(ctx, req.NamespacedName, &syncConfig); err != nil {
		if apierrs.IsNotFound(err) {
			log.V(1).Info("ignoring object delete")
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
	selector := syncConfig.Spec
	var err error
	var matchingProjects []harbor.Project
	var matcher *regexp.Regexp

	allProjects, err := r.Harbor.ListProjects()
	if err != nil {
		log.Error(err, "could not list harbor projects")
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 15}, err
	}
	if selector.Type != crdv1.RegexMatching {
		log.Error(fmt.Errorf("invalid selector type: %s", selector.Type), "selector type must be regex")
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 15}, err
	}
	matcher, err = regexp.Compile(selector.ProjectName)
	if err != nil {
		log.Error(err, "error compiling regex", "selector_project_name", selector.ProjectName)
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 15}, err
	}
	for _, project := range allProjects {
		if matcher.MatchString(project.Name) {
			log.V(1).Info("project match", "type", selector.Type, "project_name", project.Name)
			matchingProjects = append(matchingProjects, project)
		}
	}
	log.V(1).Info("found matching projects", "matching_projects", len(matchingProjects), "all_projects", len(allProjects))
	matchingProjectsGauge.WithLabelValues(syncConfig.ObjectMeta.Name, string(selector.Type), selector.ProjectName).Set(float64(len(matchingProjects)))

	// check if projects have a specific robot account
	// create it if not
	for _, project := range matchingProjects {
		projectCredential, changed, err := reconcileRobotAccounts(r.Harbor, log.WithName("reconcile_robots"), &syncConfig, project, selector.RobotAccountSuffix)
		if err != nil {
			log.Error(err, "error reconciling robot accounts")
			continue
		}

		if changed {
			log.Info("robot account changed. sending webhook", "robot_name", projectCredential.Name)
			payload := crdv1.WebhookUpdatePayload{
				Project:     project.Name,
				Credentials: *projectCredential,
			}
			data, err := json.Marshal(payload)
			if err != nil {
				log.Error(err, "failed to encode webhook payload")
			}
			for _, wh := range syncConfig.Spec.Webhook {
				u, err := url.Parse(wh.Endpoint)
				if err != nil {
					log.Error(err, "failed to parse webhook url", "url", wh.Endpoint)
					continue
				}
				req, err := http.NewRequest("POST", u.String(), bytes.NewReader(data))
				res, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Error(err, "error sending webhook", "url", wh.Endpoint)
					continue
				}
				if res.StatusCode > 300 {
					log.Error(fmt.Errorf("unexpected http response code: %d", res.StatusCode), "url", wh.Endpoint)
					continue
				}
				log.Info("successfully sent webhook", "url", wh.Endpoint)
			}
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

	err = r.Update(context.Background(), &syncConfig)
	if err != nil {
		log.Error(err, "could not update syncConfig status field", "sync_config_name", syncConfig.Name)
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 15}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager setup the controller with the manager and the event input channel
// the input chan is used to trigger recon based on external events (harbor API resources changed, forced sync)
func (r *HarborSyncConfigReconciler) SetupWithManager(mgr ctrl.Manager, input <-chan event.GenericEvent) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.HarborSync{}).
		Watches(&source.Channel{Source: input}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

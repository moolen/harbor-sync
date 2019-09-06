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
	"regexp"
	"strconv"
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

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.harborsync.k8s.io,resources=harborsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.harborsync.k8s.io,resources=harborsyncs/status,verbs=get;update;patch

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
	selector := syncConfig.Spec
	matches, err := findMatches(syncConfig, r.Harbor)
	log.V(1).Info("found matching projects", "matching_projects", len(matches))
	matchingProjectsGauge.WithLabelValues(syncConfig.ObjectMeta.Name, string(selector.Type), selector.ProjectName).Set(float64(len(matches)))

	// reconcile robot accounts
	for _, project := range matches {
		credential, changed, err := reconcileRobotAccounts(r.Harbor, log.WithName("reconcile_robots"), &syncConfig, project, selector.RobotAccountSuffix)
		if err != nil {
			log.Error(err, "error reconciling robot accounts")
			continue
		}

		if changed {
			log.Info("robot account changed. sending webhook", "robot_name", credential.Name)
			err = runWebhook(syncConfig.ObjectMeta.Name, syncConfig.Spec.Webhook, project, credential)
			if err != nil {
				log.Error(err, "error calling webhook")
			}
		}

		// reconcile secrets in namespaces
		for _, mapping := range selector.Mapping {
			f, err := mappingFuncForConfig(mapping)
			if err != nil {
				log.Error(err, "failed to get mapping for config")
				continue
			}
			err = f(r, mapping, syncConfig, project, *credential, r.Harbor.BaseURL())
			if err != nil {
				log.Error(err, "mapping failed")
				continue
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

// findMatches filters from a list of projects those projects that match the given syncConfig
func findMatches(syncConfig crdv1.HarborSync, api harbor.API) ([]harbor.Project, error) {
	var matchingProjects []harbor.Project
	allProjects, err := api.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("could not list harbor projects: %s", err.Error())
	}
	if syncConfig.Spec.Type != crdv1.RegexMatching {
		return nil, fmt.Errorf("invalid selector type: %s", syncConfig.Spec.Type)
	}
	matcher, err := regexp.Compile(syncConfig.Spec.ProjectName)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex: %s", err.Error())
	}
	for _, project := range allProjects {
		if matcher.MatchString(project.Name) {
			matchingProjects = append(matchingProjects, project)
		}
	}
	matchingProjectsGauge.WithLabelValues(syncConfig.ObjectMeta.Name, string(syncConfig.Spec.Type), syncConfig.Spec.ProjectName).Set(float64(len(matchingProjects)))
	return matchingProjects, nil
}

// runWebhook issues HTTP Requests for the configured webhooks
func runWebhook(syncConfigName string, webhookCfg []crdv1.WebhookConfig, project harbor.Project, credential *crdv1.RobotAccountCredential) error {
	payload := crdv1.WebhookUpdatePayload{
		Project:     project.Name,
		Credentials: *credential,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode webhook payload: %s", err.Error())
	}
	var errs []string
	for _, wh := range webhookCfg {
		req, err := http.NewRequest("POST", wh.Endpoint, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			webhookCounter.WithLabelValues(syncConfigName, wh.Endpoint, "error").Inc()
			errs = append(errs, fmt.Sprintf("error sending webhook to %s: %s", wh.Endpoint, err.Error()))
			continue
		}
		webhookCounter.WithLabelValues(syncConfigName, wh.Endpoint, strconv.Itoa(res.StatusCode)).Inc()
		if res.StatusCode > 300 {
			errs = append(errs, fmt.Sprintf("unexpected http response code: %d for %s", res.StatusCode, wh.Endpoint))
			continue
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("webhook errors: %#v", errs)
}

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

package main

import (
	"flag"
	"os"
	"time"

	"strings"

	"github.com/blang/semver"
	"github.com/go-logr/glogr"
	"github.com/golang/glog"
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/controllers"
	"github.com/moolen/harbor-sync/pkg/harbor"
	"github.com/moolen/harbor-sync/pkg/harbor/repository"
	"github.com/moolen/harbor-sync/pkg/store"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = crdv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var harborPollInterval time.Duration
	var forceSyncInterval time.Duration
	var rotationInterval time.Duration
	var storePath string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&storePath, "store", "/data", "path to the credentials cache")
	flag.DurationVar(&harborPollInterval, "harbor-poll-interval", time.Minute*5, "poll interval to update harbor projects & robot accounts")
	flag.DurationVar(&forceSyncInterval, "force-sync-interval", time.Minute*10, "set this to force reconciliation after a certain time")
	flag.DurationVar(&rotationInterval, "rotation-interval", time.Minute*60, "set this to rotate the credentials after the specified time")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Set("logtostderr", "true")
	flag.Parse()

	harborAPIEndpoint := os.Getenv("HARBOR_API_ENDPOINT")
	harborUsername := os.Getenv("HARBOR_USERNAME")
	harborPassword := os.Getenv("HARBOR_PASSWORD")

	log := glogr.New().WithName("controller")
	ctrl.SetLogger(log)
	defer glog.Flush()

	store, err := store.New(storePath)
	if err != nil {
		setupLog.Error(err, "unable to create credential store")
		os.Exit(1)
	}

	harborClient, err := harbor.New(harborAPIEndpoint, harborUsername, harborPassword)
	if err != nil {
		setupLog.Error(err, "unable to create harbor client")
		os.Exit(1)
	}

	info, err := harborClient.SystemInfo()
	if err != nil {
		setupLog.Error(err, "unable to get harbor system info ")
		os.Exit(1)
	}

	check, _ := semver.Make("1.8.0")
	v, err := semver.Make(strings.TrimLeft(info.HarborVersion, "v"))
	if err != nil {
		setupLog.Error(err, "unable to validate harbor version")
		os.Exit(1)
	}
	v.Pre = nil // pre-releases are OK, too
	if v.LT(check) {
		setupLog.Info("your harbor version does not support robot accounts", "harbor_version", v.String(), "required_version", check.String())
		os.Exit(1)
	}

	harborRepo, err := repository.New(harborClient, ctrl.Log.WithName("repository"), harborPollInterval)
	if err != nil {
		setupLog.Error(err, "unable to create harbor repository")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// we want to force a reconciliation
	forceSyncChan := make(chan struct{})
	go func() {
		for {
			<-time.After(forceSyncInterval)
			forceSyncChan <- struct{}{}
		}
	}()

	harborProjectChanges := harborRepo.Sync()
	syncChannels := []<-chan struct{}{forceSyncChan, harborProjectChanges}
	adapter := controllers.NewAdapter(mgr.GetClient(), ctrl.Log.WithName("adapter"), syncChannels)
	syncCfgChanges := adapter.Run()

	if err = (&controllers.HarborSyncConfigReconciler{
		CredCache:        store,
		RotationInterval: rotationInterval,
		Client:           mgr.GetClient(),
		Log:              ctrl.Log.WithName("reconciler"),
		Harbor:           harborRepo,
	}).SetupWithManager(mgr, syncCfgChanges); err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

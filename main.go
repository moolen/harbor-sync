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
	"github.com/hashicorp/go-version"
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/controllers"
	"github.com/moolen/harbor-sync/pkg/harbor"
	"github.com/moolen/harbor-sync/pkg/harbor/repository"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
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
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.DurationVar(&harborPollInterval, "harbor-poll-interval", time.Minute*5, "poll interval to update harbor projects & robot accounts")
	flag.DurationVar(&forceSyncInterval, "force-sync-interval", time.Minute*10, "force reconciliation interval")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	harborAPIEndpoint := os.Getenv("HARBOR_API_ENDPOINT")
	harborUsername := os.Getenv("HARBOR_USERNAME")
	harborPassword := os.Getenv("HARBOR_PASSWORD")
	ctrl.SetLogger(zap.Logger(true))

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
	check, _ := version.NewVersion("1.8.0")
	v, err := version.NewVersion(info.HarborVersion)
	if err != nil {
		setupLog.Error(err, "unable to validate harbor version")
		os.Exit(1)
	}
	if v.LessThan(check) {
		setupLog.Info("your harbor version does not support robot accounts", "harbor_version", v, "required_version", check)
		os.Exit(1)
	}

	harborRepo, err := repository.New(harborClient, ctrl.Log.WithName("controllers").WithName("HarborRepository"), harborPollInterval)
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
	adapter := controllers.NewAdapter(mgr.GetClient(), ctrl.Log.WithName("controllers").WithName("HarborSyncAdapter"), syncChannels)
	syncCfgChanges := adapter.Run()

	if err = (&controllers.HarborSyncConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("HarborSyncConfig"),
		Harbor: harborRepo,
	}).SetupWithManager(mgr, syncCfgChanges); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HarborSyncConfig")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

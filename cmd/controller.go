package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os"

	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/blang/semver"
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/controllers"
	"github.com/moolen/harbor-sync/pkg/harbor"
	"github.com/moolen/harbor-sync/pkg/harbor/repository"
	"github.com/moolen/harbor-sync/pkg/store"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme             *runtime.Scheme
	metricsAddr        *string
	harborPollInterval time.Duration
	forceSyncInterval  time.Duration
	rotationInterval   time.Duration
	harborAPIEndpoint  *string
	harborAPIUsername  *string
	harborAPIPassword  *string
)

func init() {
	scheme = runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		log.Fatal(err)
	}
	err = crdv1.AddToScheme(scheme)
	if err != nil {
		log.Fatal(err)
	}

	flags := controllerCmd.PersistentFlags()
	harborAPIEndpoint = flags.String("harbor-api-endpoint", "", "URL to the Harbor API Endpoint")
	harborAPIUsername = flags.String("harbor-username", "", "Harbor username to use for authentication")
	harborAPIPassword = flags.String("harbor-password", "", "Harbor password to use for authentication")
	metricsAddr = flags.String("metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flags.DurationVar(&harborPollInterval, "harbor-poll-interval", time.Minute*5, "poll interval to update harbor projects & robot accounts")
	flags.DurationVar(&forceSyncInterval, "force-sync-interval", time.Minute*10, "set this to force reconciliation after a certain time")
	flags.DurationVar(&rotationInterval, "rotation-interval", time.Minute*60, "set this to rotate the credentials after the specified time")
	viper.BindPFlags(flags)
	viper.BindEnv("harbor-username", "HARBOR_USERNAME")
	viper.BindEnv("harbor-password", "HARBOR_PASSWORD")
	viper.BindEnv("harbor-api-endpoint", "HARBOR_API_ENDPOINT")
	rootCmd.AddCommand(controllerCmd)
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Controller should run inside Kubernetes. It reconciles the desired state by managing the robot accounts in Harbor.",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := store.New(storePath)
		if err != nil {
			log.Error(err, "unable to create credential store")
			os.Exit(1)
		}

		harborClient, err := harbor.New(
			viper.GetString("harbor-api-endpoint"),
			viper.GetString("harbor-username"),
			viper.GetString("harbor-password"),
		)
		if err != nil {
			log.Error(err, "unable to create harbor client")
			os.Exit(1)
		}

		err = checkHarbor(harborClient)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		harborRepo, err := repository.New(harborClient, ctrl.Log.WithName("repository"), harborPollInterval)
		if err != nil {
			log.Error(err, "unable to create harbor repository")
			os.Exit(1)
		}

		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: *metricsAddr,
		})
		if err != nil {
			log.Error(err, "unable to start manager")
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
			Harbor:           harborRepo,
		}).SetupWithManager(mgr, syncCfgChanges); err != nil {
			log.Error(err, "unable to create controller")
			os.Exit(1)
		}

		// +kubebuilder:scaffold:builder
		log.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			log.Error(err, "problem running manager")
			os.Exit(1)
		}

	},
}

func checkHarbor(client *harbor.Client) error {
	info, err := client.SystemInfo()
	if err != nil {
		return fmt.Errorf("unable to get harbor system info")
	}
	check, _ := semver.Make("1.8.0")
	v, err := semver.Make(strings.TrimLeft(info.HarborVersion, "v"))
	if err != nil {
		return fmt.Errorf("unable to validate harbor version")
	}
	v.Pre = nil // pre-releases are OK, too
	if v.LT(check) {
		return fmt.Errorf("your harbor version does not support robot accounts")
	}
	return nil
}

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os"

	log "github.com/sirupsen/logrus"

	"github.com/blang/semver"
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/controllers"
	"github.com/moolen/harbor-sync/pkg/harbor"
	"github.com/moolen/harbor-sync/pkg/harbor/repository"
	store "github.com/moolen/harbor-sync/pkg/store/crd"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme            *runtime.Scheme
	metricsAddr       *string
	harborAPIEndpoint *string
	harborAPIUsername *string
	harborAPIPassword *string
	skipVerifyTLS     *bool
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
	skipVerifyTLS = flags.Bool("skip-tls-verification", false, "Skip TLS certificate verification")
	metricsAddr = flags.String("metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flags.Duration("harbor-poll-interval", time.Minute*5, "poll interval to update harbor projects & robot accounts")
	flags.Duration("force-sync-interval", time.Minute*10, "set this to force reconciliation after a certain time")
	flags.Duration("rotation-interval", time.Minute*60, "set this to rotate the credentials after the specified time")
	flags.Bool("leader-elect", true, "enable leader election")
	flags.String("namespace", "kube-system", "namespace in which harbor-sync runs (used for leader-election)")
	viper.BindPFlags(flags)
	viper.BindEnv("harbor-username", "HARBOR_USERNAME")
	viper.BindEnv("harbor-password", "HARBOR_PASSWORD")
	viper.BindEnv("harbor-api-endpoint", "HARBOR_API_ENDPOINT")
	viper.BindEnv("leader-elect", "LEADER_ELECT")
	viper.BindEnv("namespace", "NAMESPACE")
	viper.BindEnv("harbor-poll-interval", "HARBOR_POLL_INTERVAL")
	viper.BindEnv("force-sync-interval", "FORCE_SYNC_INTERVAL")
	viper.BindEnv("rotation-interval", "ROTATION_INTERVAL")
	rootCmd.AddCommand(controllerCmd)
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Controller should run inside Kubernetes. It reconciles the desired state by managing the robot accounts in Harbor.",
	Run: func(cmd *cobra.Command, args []string) {
		// dump cfg
		log.WithFields(log.Fields{
			"harbor-api-endpoint":  viper.GetBool("harbor-api-endpoint"),
			"leader-elect":         viper.GetBool("leader-elect"),
			"loglevel":             viper.GetString("loglevel"),
			"namespace":            viper.GetDuration("namespace"),
			"force-sync-interval":  viper.GetDuration("force-sync-interval"),
			"rotation-interval":    viper.GetDuration("rotation-interval"),
			"harbor-poll-interval": viper.GetDuration("harbor-poll-interval"),
		}).Info()

		harborClient, err := harbor.New(
			viper.GetString("harbor-api-endpoint"),
			viper.GetString("harbor-username"),
			viper.GetString("harbor-password"),
			viper.GetBool("skip-tls-verification"),
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

		harborRepo, err := repository.New(harborClient, viper.GetDuration("harbor-poll-interval"))
		if err != nil {
			log.Error(err, "unable to create harbor repository")
			os.Exit(1)
		}

		leaseDuration := 100 * time.Second
		renewDeadline := 80 * time.Second
		retryPeriod := 20 * time.Second
		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: *metricsAddr,
			LeaderElection:     viper.GetBool("leader-elect"),
			LeaderElectionID:   "harbor-sync-leader-election",
			LeaseDuration:      &leaseDuration,
			RenewDeadline:      &renewDeadline,
			RetryPeriod:        &retryPeriod,
		})
		if err != nil {
			log.Error(err, "unable to start manager")
			os.Exit(1)
		}

		// we want to force reconciliation after a certain interval
		forceSyncChan := make(chan struct{})
		go func() {
			for {
				<-time.After(viper.GetDuration("force-sync-interval"))
				forceSyncChan <- struct{}{}
			}
		}()

		harborProjectChanges := harborRepo.Sync()
		syncChannels := []<-chan struct{}{forceSyncChan, harborProjectChanges}
		adapter := controllers.NewAdapter(mgr.GetClient(), syncChannels)
		syncCfgChanges := adapter.Run()

		crdStore, err := store.New(mgr.GetClient())
		if err != nil {
			log.Fatal(err, "unable to create store")
		}

		if err = (&controllers.HarborSyncConfigReconciler{
			CredCache:        crdStore,
			RotationInterval: viper.GetDuration("rotation-interval"),
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

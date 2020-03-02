package cmd

import (
	"io/ioutil"
	"os"
	"time"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/controllers"
	"github.com/moolen/harbor-sync/pkg/harbor"
	"github.com/moolen/harbor-sync/pkg/harbor/repository"
	store "github.com/moolen/harbor-sync/pkg/store/disk"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
)

func init() {
	flags := standaloneCmd.PersistentFlags()
	flags.String("store", "/data", "path in which the credentials will be stored")
	flags.String("config", "", "path to the config file which contains the mapping. This file should be a manifest of kind: HarborSync")
	viper.BindEnv("store", "STORE")
	controllerCmd.AddCommand(standaloneCmd)
}

var standaloneCmd = &cobra.Command{
	Use:   "standalone",
	Short: "Runs the controller in standalone mode. Does not require Kubernetes. It manages robot accounts and sends webhooks.",
	Run: func(cmd *cobra.Command, args []string) {
		//
		store, err := store.New(viper.GetString("store"))
		if err != nil {
			log.Error(err, "unable to create credential store")
			os.Exit(1)
		}

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

		data, err := ioutil.ReadFile(viper.GetString("config"))
		if err != nil {
			log.Error(err, "unable to read config path")
			os.Exit(1)
		}

		codec := serializer.NewCodecFactory(scheme)
		obj, _, err := codec.UniversalDeserializer().Decode(data, nil, nil)
		if err != nil {
			log.Error(err, "unable to decode config")
			os.Exit(1)
		}
		syncConfig := obj.(*crdv1.HarborSync)
		<-harborRepo.Sync()

		for {
			err = controllers.Reconcile(*syncConfig, harborRepo, store, viper.GetDuration("rotation-interval"), nil)
			if err != nil {
				log.Error(err, "error reconciling")
			}
			log.Info("done recon")
			<-time.After(viper.GetDuration("force-sync-interval"))
		}
	},
}

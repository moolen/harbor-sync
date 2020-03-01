package cmd

import (
	"os"

	store "github.com/moolen/harbor-sync/pkg/store/disk"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {

	storeCmd.AddCommand(listCmd)
	rootCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Store subcommand lets you interact with the credential store",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Prints the store information",
	Long:  `The store contains the robot accounts. This command outputs the contents of the store.`,
	Run: func(cmd *cobra.Command, args []string) {
		store, err := store.New(viper.GetString("store"))
		if err != nil {
			log.Errorf("unable to create credential store: %s", err)
			os.Exit(1)
		}
		keys := store.Keys()
		log.Info("found items", "count", len(keys))
		for _, pair := range keys {
			cred, err := store.Get(pair[0], pair[1])
			if err != nil {
				log.WithFields(log.Fields{
					"project": pair[0],
					"name":    pair[1],
				}).Errorf("unable to read from store: %s", err.Error())
				os.Exit(1)
			}
			log.WithFields(log.Fields{
				"project": pair[0],
				"robot":   pair[1],
				"user":    cred.Name,
				"token":   cred.Token,
			}).Info("credential")
		}
	},
}

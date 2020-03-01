package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "harbor-sync",
	Short: "Harbor Sync allows you to synchronize your Harbor robot accounts with your Kubernetes cluster.",
	Long:  `Harbor Sync may run inside the kubernetes cluster (use controller subcommand) or standalone.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		lvl, err := log.ParseLevel(viper.GetString("loglevel"))
		if err != nil {
			log.Fatal(err)
		}
		log.SetLevel(lvl)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	viper.AutomaticEnv()
	flags := rootCmd.PersistentFlags()
	flags.String("loglevel", "info", "set the loglevel")
	flags.String("store", "/data", "path in which the credentials will be stored")
	viper.BindPFlags(flags)
	viper.BindEnv("loglevel", "LOGLEVEL")
	viper.BindEnv("store", "STORE")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

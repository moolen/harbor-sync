package main

import (
	"flag"
	"os"

	"github.com/go-logr/glogr"
	"github.com/golang/glog"
	"github.com/moolen/harbor-sync/pkg/store"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	storePath string
	log       = ctrl.Log.WithName("setup")
)

func main() {
	flag.StringVar(&storePath, "store", "/data", "path to the credentials cache")
	flag.Set("logtostderr", "true")
	flag.Parse()

	log := glogr.New().WithName("controller")
	ctrl.SetLogger(log)
	defer glog.Flush()

	store, err := store.New(storePath)
	if err != nil {
		log.Error(err, "unable to create credential store")
		os.Exit(1)
	}

	keys := store.Keys()
	log.Info("found items", "count", len(keys))
	for i, pair := range keys {
		cred, err := store.Get(pair[0], pair[1])
		if err != nil {
			log.Error(err, "unable to read from store", "project", pair[0], "name", pair[1])
			os.Exit(1)
		}
		log.Info(string(i), "project", pair[0], "robot", pair[1], "user", cred.Name, "token", cred.Token)
	}
}

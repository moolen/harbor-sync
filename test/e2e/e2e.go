package e2e

import (
	"testing"

	"github.com/onsi/ginkgo"
	gconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	// required
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/moolen/harbor-sync/test/e2e/framework"
)

var (
	cl client.Client
)

// RunE2ETests runs e2e tests using the Ginkgo runner
func RunE2ETests(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	var err error
	log.SetOutput(ginkgo.GinkgoWriter)
	err = crdv1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal("unable to register crd", err)
	}
	cl, err = client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		log.Fatal("unable to create client", err)
	}
	log.Infof("Starting e2e run %q on Ginkgo node %d", framework.RunID, gconfig.GinkgoConfig.ParallelNode)
	ginkgo.RunSpecs(t, "harbor-sync e2e suite")
}

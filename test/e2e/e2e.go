package e2e

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	log "github.com/sirupsen/logrus"

	// required
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/moolen/harbor-sync/test/e2e/framework"
)

// RunE2ETests runs e2e tests using the Ginkgo runner
func RunE2ETests(t *testing.T) {
	log.SetOutput(ginkgo.GinkgoWriter)
	log.Infof("Starting e2e run %q on Ginkgo node %d", framework.RunID, config.GinkgoConfig.ParallelNode)
	ginkgo.RunSpecs(t, "harbor-sync e2e suite")
}

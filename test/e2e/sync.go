package e2e

import (
	"fmt"

	"github.com/stretchr/testify/assert"

	"github.com/moolen/harbor-sync/pkg/harbor"
	"github.com/moolen/harbor-sync/test/e2e/framework"
	"github.com/onsi/ginkgo"
	log "github.com/sirupsen/logrus"
)

var _ = ginkgo.Describe("[Sync]", func() {
	f := framework.NewDefaultFramework("log")

	ginkgo.BeforeEach(func() {
		f.Harbor.EnsureUpdate(f.Namespace, harbor.SystemInfoResponse{
			HarborVersion: "1.10.1",
		}, []harbor.Project{
			{
				Name: fmt.Sprintf("proj-%s-foo", f.Namespace),
				ID:   0,
			},
			{
				Name: fmt.Sprintf("proj-%s-bar", f.Namespace),
				ID:   100,
			},
		}, map[string][]harbor.Robot{})
		_, err := framework.CreateNamespace(fmt.Sprintf("team-%s-foo", f.Namespace), f.KubeClientSet)
		assert.Nil(ginkgo.GinkgoT(), err, "error creating team namespace")

		_, err = framework.CreateNamespace(fmt.Sprintf("team-%s-bar", f.Namespace), f.KubeClientSet)
		assert.Nil(ginkgo.GinkgoT(), err, "error creating team namespace")

	})

	ginkgo.AfterEach(func() {
		err := framework.DeleteKubeNamespace(f.KubeClientSet, fmt.Sprintf("team-%s-foo", f.Namespace))
		assert.Nil(ginkgo.GinkgoT(), err, "error creating team namespace")

		err = framework.DeleteKubeNamespace(f.KubeClientSet, fmt.Sprintf("team-%s-bar", f.Namespace))
		assert.Nil(ginkgo.GinkgoT(), err, "error creating team namespace")

	})

	ginkgo.It("should do sync robot accounts", func() {
		log.Infof("foo: %s", f.BaseName)
		err := framework.WaitForSecret(
			f.KubeClientSet,
			framework.DefaultTimeout,
			fmt.Sprintf("proj-%s-foo-pull-secret", f.Namespace),
			fmt.Sprintf("team-%s-foo", f.Namespace), map[string]string{})
		assert.Nil(ginkgo.GinkgoT(), err, "error syncing robot account")
		err = framework.WaitForSecret(
			f.KubeClientSet,
			framework.DefaultTimeout,
			fmt.Sprintf("proj-%s-bar-pull-secret", f.Namespace),
			fmt.Sprintf("team-%s-bar", f.Namespace), map[string]string{})
		assert.Nil(ginkgo.GinkgoT(), err, "error syncing robot account")
	})
})

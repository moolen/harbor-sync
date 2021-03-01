package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/credentialprovider"

	"github.com/stretchr/testify/assert"

	"github.com/moolen/harbor-sync/test/e2e/framework"
	"github.com/onsi/ginkgo"
)

var _ = ginkgo.Describe("[Sync]", func() {
	f := framework.NewDefaultFramework("sync")

	ginkgo.BeforeEach(func() {
		f.EnsureProjects([]string{
			fmt.Sprintf("proj-%s-foo", f.Namespace),
			fmt.Sprintf("proj-%s-bar", f.Namespace),
		})
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

		for _, row := range []struct {
			name      string
			namespace string
		}{
			{
				name:      fmt.Sprintf("proj-%s-foo-pull-secret", f.Namespace),
				namespace: fmt.Sprintf("team-%s-foo", f.Namespace),
			},
			{
				name:      fmt.Sprintf("proj-%s-bar-pull-secret", f.Namespace),
				namespace: fmt.Sprintf("team-%s-bar", f.Namespace),
			},
		} {
			err := framework.WaitForSecret(
				f.KubeClientSet,
				framework.DefaultTimeout,
				row.name,
				row.namespace,
				map[string]string{})
			assert.Nil(ginkgo.GinkgoT(), err, "error syncing robot account")
			secret, err := f.KubeClientSet.CoreV1().Secrets(row.namespace).Get(context.Background(), row.name, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err, "getting secrets")
			assert.Contains(ginkgo.GinkgoT(), secret.Data, ".dockerconfigjson")
			credData := secret.Data[".dockerconfigjson"]
			var cred credentialprovider.DockerConfigJSON
			err = json.Unmarshal(credData, &cred)

			for u, auth := range cred.Auths {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/service/token", u), nil)
				assert.Nil(ginkgo.GinkgoT(), err, "building  secrets")
				q := req.URL.Query()
				q.Set("account", auth.Username)
				q.Set("client_id", "docker")
				q.Set("offline_token", "true")
				q.Set("service", "harbor-registry")
				req.URL.RawQuery = q.Encode()
				req.SetBasicAuth(auth.Username, auth.Password)
				ob, _ := httputil.DumpRequestOut(req, true)
				fmt.Fprintln(ginkgo.GinkgoWriter, string(ob))
				res, err := http.DefaultClient.Do(req)
				assert.Nil(ginkgo.GinkgoT(), err, "sending request")
				assert.Equal(ginkgo.GinkgoT(), http.StatusOK, res.StatusCode)
				res.Body.Close()
			}
		}
	})
})

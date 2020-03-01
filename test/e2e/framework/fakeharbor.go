package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/moolen/harbor-sync/pkg/harbor"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

// FakeHarbor ..
type FakeHarbor struct {
	URL string
}

// EnsureUpdate ..
func (f *FakeHarbor) EnsureUpdate(namespace string, info harbor.SystemInfoResponse, projects []harbor.Project, robots map[string][]harbor.Robot) {
	update := func(target string, data interface{}) {
		buf := bytes.NewBuffer(nil)
		err := json.NewEncoder(buf).Encode(data)
		assert.Nil(ginkgo.GinkgoT(), err, "update harbor")
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://fake-harbor-api.%s.svc.cluster.local/_update/%s", namespace, target), buf)
		assert.Nil(ginkgo.GinkgoT(), err, "update harbor")
		res, err := http.DefaultClient.Do(req)
		assert.Nil(ginkgo.GinkgoT(), err, "update harbor")
		assert.Equal(ginkgo.GinkgoT(), http.StatusOK, res.StatusCode)
	}
	update("systeminfo", info)
	update("projects", projects)
	update("robots", robots)
}

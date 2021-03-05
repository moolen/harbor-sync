package framework

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mittwald/goharbor-client/apiv1/project"
	harborclient "github.com/mittwald/goharbor-client/v3/apiv2"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/credentialprovider"

	"k8s.io/client-go/tools/clientcmd"
)

var (
	// KubectlPath defines the full path of the kubectl binary
	KubectlPath = "/usr/local/bin/kubectl"
)

// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
type Framework struct {
	BaseName string

	// A Kubernetes and Service Catalog client
	KubeClientSet          kubernetes.Interface
	KubeConfig             *restclient.Config
	APIExtensionsClientSet apiextcs.Interface

	Harbor *harborclient.RESTClient

	Namespace string
}

const (
	harborService  = "harbor.default.svc.cluster.local"
	harborUser     = "admin"
	harborPassword = "Harbor12345"
)

// NewDefaultFramework makes a new framework and sets up a BeforeEach/AfterEach for
// you (you can write additional before/after each functions).
func NewDefaultFramework(baseName string) *Framework {
	defer ginkgo.GinkgoRecover()

	var kubeConfig *restclient.Config
	var err error
	kcPath := os.Getenv("KUBECONFIG")
	if kcPath != "" {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kcPath)
	} else {
		kubeConfig, err = restclient.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}
	assert.Nil(ginkgo.GinkgoT(), err, "creting kubernetes API client configuration")

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	assert.Nil(ginkgo.GinkgoT(), err, "creating Kubernetes API client")

	harborClient, err := harborclient.NewRESTClientForHost(
		fmt.Sprintf("https://%s/api", harborService), harborUser, harborPassword)
	assert.Nil(ginkgo.GinkgoT(), err, "creating harbor client")

	f := &Framework{
		BaseName:      baseName,
		KubeConfig:    kubeConfig,
		KubeClientSet: kubeClient,
		Harbor:        harborClient,
	}
	ginkgo.BeforeEach(f.BeforeEach)
	ginkgo.AfterEach(f.AfterEach)
	return f
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	ns, err := CreateKubeNamespace(f.BaseName, f.KubeClientSet)
	assert.Nil(ginkgo.GinkgoT(), err, "creating namespace")
	f.Namespace = ns

	log.Infof("creating harbor-sync controller")
	err = f.createHarborSyncController(f.Namespace)
	assert.Nil(ginkgo.GinkgoT(), err, "creating harbor-sync")
}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	if ginkgo.CurrentGinkgoTestDescription().Failed {
		// spit-out logs and what not
	}

	err := DeleteKubeNamespace(f.KubeClientSet, f.Namespace)
	assert.Nil(ginkgo.GinkgoT(), err, "deleting namespace %v", f.Namespace)
}

// EnsureProjects creates projects if they do not exist yet
func (f *Framework) EnsureProjects(projects []string) {
	for _, projectName := range projects {
		p, err := f.Harbor.GetProjectByName(context.Background(), projectName)
		if err != nil {
			if err.Error() != project.ErrProjectNotFoundMsg {
				log.Errorf("unable to get project: %s", err)
				assert.Nil(ginkgo.GinkgoT(), err, "get fetch %s", projectName)
			}
		}
		if p != nil {
			log.Infof("project already exists: %s", p.Name)
			continue
		}
		log.Infof("creating project %s", projectName)
		p, err = f.Harbor.NewProject(context.Background(), projectName, nil)
		if err != nil {
			assert.Nil(ginkgo.GinkgoT(), err, "create project %s", projectName)
		}
		log.Infof("created project %s / id %d", p.Name, p.ProjectID)
	}
}

// EnsureImages ...
func (f *Framework) EnsureImages(projectImageMap map[string][]string) {
	out, err := exec.Command(
		"crane",
		"auth",
		"login",
		harborService,
		"-u", harborUser,
		"-p", harborPassword,
	).CombinedOutput()
	log.Infof("out: %s", out)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	for project, images := range projectImageMap {
		for _, image := range images {
			log.Infof("pushing image %s into project %s", image, project)
			out, err := exec.Command(
				"crane",
				"cp",
				image,
				fmt.Sprintf("%s/%s/%s", harborService, project, image),
			).CombinedOutput()
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "stdout: %s", out)
		}
	}
}

// TestPullSecret uses the provided credentials to pull img
func (f *Framework) TestPullSecret(creds credentialprovider.DockerConfigJSON, img string) {
	for _, auth := range creds.Auths {
		log.Infof("logging in at %s as %s with %s", harborService, auth.Username, auth.Password)
		out, err := exec.Command(
			"crane",
			"auth",
			"login",
			harborService,
			"-u", auth.Username,
			"-p", auth.Password,
		).CombinedOutput()
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "stdout: %s", out)
		out, err = exec.Command(
			"crane",
			"pull",
			img,
			"example.tgz",
		).CombinedOutput()
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "stdout: %s", out)
	}
}

func (f *Framework) createHarborSyncController(namespace string) error {
	cmdRoot := os.Getenv("E2E_EXEC_ROOT")
	cmd := exec.Command(filepath.Join(cmdRoot, "/deploy.sh"), namespace)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unexpected error waiting for harbor-sync deployment: %v.\nLogs:\n%v", err, string(out))
	}
	return nil
}

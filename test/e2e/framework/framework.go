package framework

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/onsi/ginkgo"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
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

	Harbor *FakeHarbor

	Namespace string
}

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

	f := &Framework{
		BaseName:      baseName,
		KubeConfig:    kubeConfig,
		KubeClientSet: kubeClient,
		Harbor:        &FakeHarbor{},
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

	err = WaitForPodsReady(f.KubeClientSet, DefaultTimeout, 2, f.Namespace, metav1.ListOptions{
		LabelSelector: "app=harbor-sync",
	})
	assert.Nil(ginkgo.GinkgoT(), err, "waiting for harbor-sync pods to be ready")

	err = WaitForPodsReady(f.KubeClientSet, DefaultTimeout, 1, f.Namespace, metav1.ListOptions{
		LabelSelector: "app=fake-harbor-api",
	})
	assert.Nil(ginkgo.GinkgoT(), err, "waiting for fake-harbor-api pods to be ready")
}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	if ginkgo.CurrentGinkgoTestDescription().Failed {
		// spit-out logs and what not
	}

	err := DeleteKubeNamespace(f.KubeClientSet, f.Namespace)
	assert.Nil(ginkgo.GinkgoT(), err, "deleting namespace %v", f.Namespace)
}

func (f *Framework) createHarborSyncController(namespace string) error {
	cmd := exec.Command("/deploy.sh", namespace)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unexpected error waiting for harbor-sync deployment: %v.\nLogs:\n%v", err, string(out))
	}

	return nil

}

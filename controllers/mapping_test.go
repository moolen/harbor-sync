package controllers

import (
	"context"
	"regexp"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
	"k8s.io/api/core/v1"

	harborfake "github.com/moolen/harbor-sync/pkg/harbor/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Mapping", func() {

	BeforeEach(func() {
		ensureNamespace(k8sClient, "team-a")
		ensureNamespace(k8sClient, "team-b")
		ensureNamespace(k8sClient, "kube-system")
	})

	AfterEach(func() {
		deleteNamespace(k8sClient, "team-a")
		deleteNamespace(k8sClient, "team-b")
		deleteNamespace(k8sClient, "kube-system")

		deleteHarborSyncConfig(k8sClient, "my-cfg")
	})

	fakeHarbor := harborfake.Client{}
	log := zap.Logger(false)
	hscr := HarborSyncConfigReconciler{
		k8sClient,
		log,
		fakeHarbor,
	}

	Describe("Match", func() {
		It("should create secrets in namespace", func(done Done) {
			var err error
			mapping := crdv1.ProjectMapping{
				Type:      crdv1.MatchMappingType,
				Namespace: "team-.*",
				Secret:    "platform-pull-token",
			}
			ensureHarborSyncConfigWithParams(k8sClient, "my-cfg", "platform-team", mapping)
			hscr.mapByMatching(
				mapping,
				regexp.MustCompile("platform-team"),
				harbor.Project{
					ID:   1,
					Name: "platform-team",
				},
				crdv1.RobotAccountCredential{
					Name:  "robot$sync-bot",
					Token: "my-token",
				},
			)

			teamASecret := v1.Secret{}
			teamBSecret := v1.Secret{}
			SystemSecret := v1.Secret{}
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-a", Name: "platform-pull-token"}, &teamASecret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-b", Name: "platform-pull-token"}, &teamBSecret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "kube-system", Name: "platform-pull-token"}, &teamBSecret)
			Expect(err).To(HaveOccurred())

			Expect(string(teamASecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(string(teamBSecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(SystemSecret.Data).To(BeNil())

			close(done)
		})
	})

	Describe("Translate", func() {
		It("should create secrets in namespace", func(done Done) {
			var err error
			mapping := crdv1.ProjectMapping{
				Type:      crdv1.MatchMappingType,
				Namespace: "team-$1",
				Secret:    "team-$1-pull-token",
			}
			ensureHarborSyncConfigWithParams(k8sClient, "my-cfg", "team-(.*)", mapping)

			hscr.mapByTranslating(
				mapping,
				regexp.MustCompile("team-(.*)"),
				harbor.Project{
					ID:   1,
					Name: "team-a",
				},
				crdv1.RobotAccountCredential{
					Name:  "robot$sync-bot",
					Token: "my-token",
				},
			)

			teamASecret := v1.Secret{}
			teamBSecret := v1.Secret{}
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-a", Name: "team-a-pull-token"}, &teamASecret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-b", Name: "team--b-pull-token"}, &teamBSecret)
			Expect(err).To(HaveOccurred())

			Expect(string(teamASecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(teamBSecret.Data).To(BeNil())

			close(done)
		})
	})

})

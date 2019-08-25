package controllers

import (
	"context"
	"regexp"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	harborfake "github.com/moolen/harbor-sync/pkg/harbor/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Mapping", func() {

	Describe("Match", func() {
		It("should create secrets in namespace", func(done Done) {
			var cl client.Client
			log := zap.Logger(true)
			scheme := runtime.NewScheme()
			mapping := crdv1.ProjectMapping{
				Type:      crdv1.MatchMappingType,
				Namespace: "team-.*",
				Secret:    "platform-pull-token",
			}
			syncConfig := crdv1.HarborSyncConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "my-cfg", Namespace: "1"},
				Spec: crdv1.HarborSyncConfigSpec{
					ProjectSelector: []crdv1.ProjectSelector{
						{
							Type:               crdv1.RegexMatching,
							ProjectName:        "platform-team",
							RobotAccountSuffix: "sync-bot",
							Mapping: []crdv1.ProjectMapping{
								mapping,
							},
						},
					},
				},
			}
			fakeHarbor := harborfake.Client{}
			crdv1.AddToScheme(scheme)
			v1.AddToScheme(scheme)
			nsTeamA := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "team-a",
				},
			}
			nsTeamB := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "team-b",
				},
			}
			nsTeamSystem := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kube-system",
				},
			}
			cl = fake.NewFakeClientWithScheme(scheme, &syncConfig, &nsTeamA, &nsTeamB, &nsTeamSystem)

			hscr := HarborSyncConfigReconciler{
				cl,
				log,
				fakeHarbor,
			}

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
			err := cl.Get(context.Background(), types.NamespacedName{Namespace: "team-a", Name: "platform-pull-token"}, &teamASecret)
			Expect(err).ToNot(HaveOccurred())
			err = cl.Get(context.Background(), types.NamespacedName{Namespace: "team-b", Name: "platform-pull-token"}, &teamBSecret)
			Expect(err).ToNot(HaveOccurred())
			err = cl.Get(context.Background(), types.NamespacedName{Namespace: "kube-system", Name: "platform-pull-token"}, &teamBSecret)
			Expect(err).To(HaveOccurred())

			Expect(string(teamASecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(string(teamBSecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(SystemSecret.Data).To(BeNil())

			close(done)
		})
	})

	Describe("Translate", func() {
		It("should create secrets in namespace", func(done Done) {
			var cl client.Client
			log := zap.Logger(true)
			scheme := runtime.NewScheme()
			mapping := crdv1.ProjectMapping{
				Type:      crdv1.MatchMappingType,
				Namespace: "team-$1",
				Secret:    "team-$1-pull-token",
			}
			syncConfig := crdv1.HarborSyncConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "my-cfg", Namespace: "1"},
				Spec: crdv1.HarborSyncConfigSpec{
					ProjectSelector: []crdv1.ProjectSelector{
						{
							Type:               crdv1.RegexMatching,
							ProjectName:        "team-(.*)",
							RobotAccountSuffix: "sync-bot",
							Mapping: []crdv1.ProjectMapping{
								mapping,
							},
						},
					},
				},
			}
			fakeHarbor := harborfake.Client{}
			crdv1.AddToScheme(scheme)
			v1.AddToScheme(scheme)
			nsTeamA := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "team-a",
				},
			}
			nsTeamB := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "team-b",
				},
			}
			cl = fake.NewFakeClientWithScheme(scheme, &syncConfig, &nsTeamA, &nsTeamB)

			hscr := HarborSyncConfigReconciler{
				cl,
				log,
				fakeHarbor,
			}

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
			err := cl.Get(context.Background(), types.NamespacedName{Namespace: "team-a", Name: "team-a-pull-token"}, &teamASecret)
			Expect(err).ToNot(HaveOccurred())
			err = cl.Get(context.Background(), types.NamespacedName{Namespace: "team-b", Name: "team--b-pull-token"}, &teamBSecret)
			Expect(err).To(HaveOccurred())

			Expect(string(teamASecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(teamBSecret.Data).To(BeNil())

			close(done)
		})
	})

})

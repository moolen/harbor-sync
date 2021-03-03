/*
Copyright 2019 The Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
	harborfake "github.com/moolen/harbor-sync/pkg/harbor/fake"
	store "github.com/moolen/harbor-sync/pkg/store/disk"
	"github.com/moolen/harbor-sync/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

const defaultRobotSecretData = `{"auths":{"":{"username":"robot$sync-bot","password":"1234","auth":"cm9ib3Qkc3luYy1ib3Q6MTIzNA=="}}}`

var _ = Describe("Controller", func() {

	var fakeHarbor *harborfake.Client
	var hscr *HarborSyncConfigReconciler
	var credStore *store.Store

	BeforeEach(func() {
		fakeHarbor = &harborfake.Client{}
		credStore, _ = store.NewTemp()
		hscr = &HarborSyncConfigReconciler{
			k8sClient,
			time.Hour * 24,
			credStore,
			fakeHarbor,
		}
	})

	AfterEach(func() {
		credStore.Reset()
	})

	Describe("Reconcile", func() {

		listProjectsResponse := []harbor.Project{
			{
				ID:   1,
				Name: "team-foo",
			},
			{
				ID:   2,
				Name: "team-bar",
			},
		}
		getRobotAccountsResponse := []harbor.Robot{
			{
				Name: "robot$sync-bot",
				// expires right now
				ExpiresAt: time.Now().Unix(),
			},
		}

		BeforeEach(func() {
			fakeHarbor.CreateRobotAccountFunc = func(name string, pushAccess bool, project harbor.Project) (*harbor.CreateRobotResponse, error) {
				return &harbor.CreateRobotResponse{
					Name:  "robot$sync-bot",
					Token: "1234",
				}, nil
			}
			fakeHarbor.ListProjectsFunc = func() ([]harbor.Project, error) {
				return listProjectsResponse, nil
			}
			fakeHarbor.GetRobotAccountsFunc = func(project harbor.Project) ([]harbor.Robot, error) {
				return getRobotAccountsResponse, nil
			}
		})

		It("should reconcile robot accounts by matching", func() {
			test.EnsureNamespace(k8sClient, "team-rcm-foo")
			test.EnsureNamespace(k8sClient, "team-rcm-bar")
			defer test.DeleteNamespace(k8sClient, "team-rcm-foo")
			defer test.DeleteNamespace(k8sClient, "team-rcm-bar")

			test.EnsureHarborSyncConfigWithParams(k8sClient, "my-rcm-cfg", "team-(.*)", &crdv1.ProjectMapping{
				Namespace: "team-rcm-.*",
				Secret:    "$1-pull-secret",
				Type:      crdv1.MatchMappingType,
			}, nil)
			defer test.DeleteHarborSyncConfig(k8sClient, "my-rcm-cfg")

			_, err := hscr.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: "my-rcm-cfg",
				},
			})

			Expect(err).ToNot(HaveOccurred())

			// expect the following secrets:
			expected := []struct {
				ns     string
				secret string
			}{
				{
					ns:     "team-rcm-foo",
					secret: "foo-pull-secret",
				},
				{
					ns:     "team-rcm-foo",
					secret: "bar-pull-secret",
				},
				{
					ns:     "team-rcm-bar",
					secret: "foo-pull-secret",
				},
				{
					ns:     "team-rcm-bar",
					secret: "bar-pull-secret",
				},
			}

			Eventually(func() bool {
				for _, expect := range expected {
					var secret v1.Secret
					err := k8sClient.Get(context.Background(), types.NamespacedName{
						Namespace: expect.ns,
						Name:      expect.secret,
					}, &secret)
					if err != nil {
						return false
					}
					if secret.Data[v1.DockerConfigJsonKey] == nil {
						return false
					}
					if string(secret.Data[v1.DockerConfigJsonKey]) != defaultRobotSecretData {
						return false
					}
				}
				return true
			}, time.Second*15, time.Second).Should(BeTrue())

			// check status spec
			var hs crdv1.HarborSync
			Eventually(func() bool {
				err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "my-rcm-cfg"}, &hs)
				if errors.IsNotFound(err) {
					return false
				}
				c := GetSyncCondition(hs.Status, crdv1.HarborSyncReady)
				if c == nil || c.Status != v1.ConditionTrue {
					return false
				}
				if len(hs.Status.ProjectList) != 2 {
					return false
				}
				return test.CheckProjects(map[string][]string{
					"team-foo": {"team-rcm-foo", "team-rcm-bar"},
					"team-bar": {"team-rcm-bar", "team-rcm-foo"},
				}, hs.Status)
			}, time.Second*15, time.Second).Should(BeTrue())
		})

		It("should reconcile robot accounts by translating", func() {
			test.EnsureNamespace(k8sClient, "team-rt-foo")
			test.EnsureNamespace(k8sClient, "team-rt-bar")
			defer test.DeleteNamespace(k8sClient, "team-rt-foo")
			defer test.DeleteNamespace(k8sClient, "team-rt-bar")

			test.EnsureHarborSyncConfigWithParams(k8sClient, "my-rt-cfg", "team-(.*)", &crdv1.ProjectMapping{
				Namespace: "team-rt-$1",
				Secret:    "default-pull-secret",
				Type:      crdv1.TranslateMappingType,
			}, nil)
			defer test.DeleteHarborSyncConfig(k8sClient, "my-rt-cfg")
			_, err := hscr.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: "my-rt-cfg",
				},
			})

			Expect(err).ToNot(HaveOccurred())

			// expect the following secrets:
			expected := []struct {
				ns     string
				secret string
			}{
				{
					ns:     "team-rt-foo",
					secret: "default-pull-secret",
				},
				{
					ns:     "team-rt-bar",
					secret: "default-pull-secret",
				},
			}

			// check that secrets get created
			Eventually(func() bool {
				for _, expect := range expected {
					var secret v1.Secret
					err := k8sClient.Get(context.Background(), types.NamespacedName{
						Namespace: expect.ns,
						Name:      expect.secret,
					}, &secret)
					Expect(err).ToNot(HaveOccurred())
					Expect(secret.Data[v1.DockerConfigJsonKey]).ToNot(BeNil())
					Expect(string(secret.Data[v1.DockerConfigJsonKey])).To(Equal(defaultRobotSecretData))
				}
				return true
			})

			// check status spec
			var hs crdv1.HarborSync
			Eventually(func() bool {
				err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "my-rt-cfg"}, &hs)
				if errors.IsNotFound(err) {
					return false
				}
				c := GetSyncCondition(hs.Status, crdv1.HarborSyncReady)
				if c == nil || c.Status != v1.ConditionTrue {
					return false
				}
				if len(hs.Status.ProjectList) != 2 {
					return false
				}
				return test.CheckProjects(map[string][]string{
					"team-foo": {"team-rt-foo"},
					"team-bar": {"team-rt-bar"},
				}, hs.Status)
			}, time.Second*15, time.Second).Should(BeTrue())
		})

		It("should call the webhook", func() {
			var fooWebhookCalled bool
			var barWebhookCalled bool

			// our webhook receiver
			srv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				// ensure we can decode the message
				var msg crdv1.WebhookUpdatePayload
				Expect(req.Header.Get("content-type")).To(Equal("application/json"))
				body, err := ioutil.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				err = json.Unmarshal(body, &msg)
				Expect(err).ToNot(HaveOccurred())
				res.WriteHeader(http.StatusOK)

				if msg.Project == "team-foo" {
					Expect(msg.Credentials.Name).To(Equal("robot$sync-bot"))
					Expect(msg.Credentials.Token).To(Equal("1234"))
					fooWebhookCalled = true
					return
				} else if msg.Project == "team-bar" {
					Expect(msg.Credentials.Name).To(Equal("robot$sync-bot"))
					Expect(msg.Credentials.Token).To(Equal("1234"))
					barWebhookCalled = true
					return
				}
				Fail("unexpected webhook payload")
			}))
			test.EnsureNamespace(k8sClient, "team-wh-foo")
			defer test.DeleteNamespace(k8sClient, "team-wh-foo")
			test.EnsureHarborSyncConfigWithParams(k8sClient, "my-wh-cfg", "team-(.*)", &crdv1.ProjectMapping{
				Namespace: "team-recon-$1",
				Secret:    "default-pull-secret",
				Type:      crdv1.TranslateMappingType,
			}, []crdv1.WebhookConfig{
				{
					Endpoint: srv.URL,
				},
			})
			defer test.DeleteHarborSyncConfig(k8sClient, "my-wh-cfg")
			_, err := hscr.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: "my-wh-cfg",
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(fooWebhookCalled).To(BeTrue())
			Expect(barWebhookCalled).To(BeTrue())
		})

		It("should call the webhook even without mapping", func() {
			var fooWebhookCalled bool
			var barWebhookCalled bool

			// our webhook receiver
			srv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				// ensure we can decode the message
				var msg crdv1.WebhookUpdatePayload
				Expect(req.Header.Get("content-type")).To(Equal("application/json"))
				body, err := ioutil.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				err = json.Unmarshal(body, &msg)
				Expect(err).ToNot(HaveOccurred())
				res.WriteHeader(http.StatusOK)

				if msg.Project == "team-foo" {
					Expect(msg.Credentials.Name).To(Equal("robot$sync-bot"))
					Expect(msg.Credentials.Token).To(Equal("1234"))
					fooWebhookCalled = true
					return
				} else if msg.Project == "team-bar" {
					Expect(msg.Credentials.Name).To(Equal("robot$sync-bot"))
					Expect(msg.Credentials.Token).To(Equal("1234"))
					barWebhookCalled = true
					return
				}
				Fail("unexpected webhook payload")
			}))
			test.EnsureNamespace(k8sClient, "team-whmp-foo")
			defer test.DeleteNamespace(k8sClient, "team-whmp-foo")
			test.EnsureHarborSyncConfigWithParams(k8sClient, "my-whmp-cfg", "team-(.*)", nil, []crdv1.WebhookConfig{
				{
					Endpoint: srv.URL,
				},
			})
			defer test.DeleteHarborSyncConfig(k8sClient, "my-whmp-cfg")
			_, err := hscr.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: "my-whmp-cfg",
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(fooWebhookCalled).To(BeTrue())
			Expect(barWebhookCalled).To(BeTrue())
		})
	})
})

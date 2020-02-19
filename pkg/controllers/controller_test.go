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
	"github.com/moolen/harbor-sync/pkg/store"
	"github.com/moolen/harbor-sync/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

const defaultRobotSecretData = `{"auths":{"":{"username":"robot$sync-bot","password":"1234","auth":"cm9ib3Qkc3luYy1ib3Q6MTIzNA=="}}}`

var _ = Describe("Controller", func() {

	var fakeHarbor *harborfake.Client
	var hscr *HarborSyncConfigReconciler
	var credStore *store.DiskStore

	BeforeEach(func() {
		test.EnsureNamespace(k8sClient, "team-recon-foo")
		test.EnsureNamespace(k8sClient, "team-recon-bar")
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
		test.DeleteNamespace(k8sClient, "team-recon-foo")
		test.DeleteNamespace(k8sClient, "team-recon-bar")
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

		AfterEach(func() {
			test.DeleteHarborSyncConfig(k8sClient, "my-recon-cfg")
		})

		It("should reconcile robot accounts by matching", func(done Done) {
			test.EnsureHarborSyncConfigWithParams(k8sClient, "my-recon-cfg", "team-(.*)", &crdv1.ProjectMapping{
				Namespace: "team-recon-.*",
				Secret:    "$1-pull-secret",
				Type:      crdv1.MatchMappingType,
			}, nil)
			_, err := hscr.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: "my-recon-cfg",
				},
			})

			Expect(err).ToNot(HaveOccurred())

			// expect the following secrets:
			expected := []struct {
				ns     string
				secret string
			}{
				{
					ns:     "team-recon-foo",
					secret: "foo-pull-secret",
				},
				{
					ns:     "team-recon-foo",
					secret: "bar-pull-secret",
				},
				{
					ns:     "team-recon-bar",
					secret: "foo-pull-secret",
				},
				{
					ns:     "team-recon-bar",
					secret: "bar-pull-secret",
				},
			}

			for _, expect := range expected {
				var secret v1.Secret
				err := k8sClient.Get(context.Background(), types.NamespacedName{
					Namespace: expect.ns,
					Name:      expect.secret,
				}, &secret)
				Expect(err).ToNot(HaveOccurred())
				Expect(secret.Data[v1.DockerConfigJsonKey]).ToNot(BeNil())
				Expect(string(secret.Data[v1.DockerConfigJsonKey])).To(Equal(defaultRobotSecretData))

				test.DeleteSecret(k8sClient, expect.ns, expect.secret)
			}

			close(done)
		})

		It("should reconcile robot accounts by translating", func(done Done) {
			test.EnsureHarborSyncConfigWithParams(k8sClient, "my-recon-cfg", "team-(.*)", &crdv1.ProjectMapping{
				Namespace: "team-recon-$1",
				Secret:    "default-pull-secret",
				Type:      crdv1.TranslateMappingType,
			}, nil)
			_, err := hscr.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: "my-recon-cfg",
				},
			})

			Expect(err).ToNot(HaveOccurred())

			// expect the following secrets:
			expected := []struct {
				ns     string
				secret string
			}{
				{
					ns:     "team-recon-foo",
					secret: "default-pull-secret",
				},
				{
					ns:     "team-recon-bar",
					secret: "default-pull-secret",
				},
			}

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
			close(done)
		})

		It("should call the webhook", func(done Done) {
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

			test.EnsureHarborSyncConfigWithParams(k8sClient, "my-recon-cfg", "team-(.*)", &crdv1.ProjectMapping{
				Namespace: "team-recon-$1",
				Secret:    "default-pull-secret",
				Type:      crdv1.TranslateMappingType,
			}, []crdv1.WebhookConfig{
				{
					Endpoint: srv.URL,
				},
			})
			_, err := hscr.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: "my-recon-cfg",
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(fooWebhookCalled).To(BeTrue())
			Expect(barWebhookCalled).To(BeTrue())
			close(done)
		})

		It("should call the webhook even without mapping", func(done Done) {
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

			test.EnsureHarborSyncConfigWithParams(k8sClient, "my-recon-cfg", "team-(.*)", nil, []crdv1.WebhookConfig{
				{
					Endpoint: srv.URL,
				},
			})
			_, err := hscr.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: "my-recon-cfg",
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(fooWebhookCalled).To(BeTrue())
			Expect(barWebhookCalled).To(BeTrue())
			close(done)
		})
	})

})

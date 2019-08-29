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
	"time"

	"github.com/go-logr/logr"
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
	harborfake "github.com/moolen/harbor-sync/pkg/harbor/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const defaultRobotSecretData = `{"auths":{"":{"username":"robot$sync-bot","password":"1234","auth":"cm9ib3Qkc3luYy1ib3Q6MTIzNA=="}}}`

var _ = Describe("Controller", func() {

	var fakeHarbor *harborfake.Client
	var log logr.Logger
	var hscr *HarborSyncConfigReconciler

	BeforeEach(func() {
		ensureNamespace(k8sClient, "team-recon-a")
		ensureNamespace(k8sClient, "team-recon-b")
		fakeHarbor = &harborfake.Client{}
		log = zap.Logger(true)
		hscr = &HarborSyncConfigReconciler{
			k8sClient,
			log,
			fakeHarbor,
		}
	})

	AfterEach(func() {
		deleteNamespace(k8sClient, "team-recon-a")
		deleteNamespace(k8sClient, "team-recon-b")
	})

	Describe("Reconcile", func() {

		var listProjectsResponse []harbor.Project
		var getRobotAccountsResponse []harbor.Robot

		BeforeEach(func() {
			fakeHarbor.CreateRobotAccountFunc = func(name string, project harbor.Project) (*harbor.CreateRobotResponse, error) {
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
			deleteHarborSyncConfig(k8sClient, "my-recon-cfg")
		})

		It("should reconcile robot accounts by matching", func(done Done) {
			listProjectsResponse = []harbor.Project{
				{
					ID:   1,
					Name: "platform-team",
				},
				{
					ID:   2,
					Name: "operations-team",
				},
			}
			getRobotAccountsResponse = []harbor.Robot{
				{
					Name: "robot$sync-bot",
					// expires right now
					ExpiresAt: int(time.Now().Unix()),
				},
			}
			ensureHarborSyncConfigWithParams(k8sClient, "my-recon-cfg", "(platform|operations)-team", crdv1.ProjectMapping{
				Namespace: "team-recon-.*",
				Secret:    "$1-pull-secret",
				Type:      crdv1.MatchMappingType,
			})
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
					ns:     "team-recon-a",
					secret: "platform-pull-secret",
				},
				{
					ns:     "team-recon-a",
					secret: "operations-pull-secret",
				},
				{
					ns:     "team-recon-b",
					secret: "platform-pull-secret",
				},
				{
					ns:     "team-recon-b",
					secret: "operations-pull-secret",
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

				deleteSecret(k8sClient, expect.ns, expect.secret)
			}

			close(done)
		})

		It("should reconcile robot accounts by translating", func(done Done) {
			listProjectsResponse = []harbor.Project{
				{
					ID:   1,
					Name: "team-a",
				},
				{
					ID:   2,
					Name: "team-b",
				},
			}
			getRobotAccountsResponse = []harbor.Robot{
				{
					Name: "robot$sync-bot",
					// expires right now
					ExpiresAt: int(time.Now().Add(time.Hour * 24).Unix()),
				},
			}
			ensureHarborSyncConfigWithParams(k8sClient, "my-recon-cfg", "team-(.*)", crdv1.ProjectMapping{
				Namespace: "team-recon-$1",
				Secret:    "default-pull-secret",
				Type:      crdv1.TranslateMappingType,
			})
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
					ns:     "team-recon-a",
					secret: "default-pull-secret",
				},
				{
					ns:     "team-recon-b",
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
	})

})

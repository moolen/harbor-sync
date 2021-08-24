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

package reconciler

import (
	"context"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
	"github.com/moolen/harbor-sync/pkg/test"
	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Mapping", func() {

	BeforeEach(func() {
	})

	Describe("Match", func() {

		BeforeEach(func() {
			test.EnsureNamespace(k8sClient, "team-match-a")
			test.EnsureNamespace(k8sClient, "team-match-b")
		})

		AfterEach(func() {
			test.DeleteNamespace(k8sClient, "team-match-a")
			test.DeleteNamespace(k8sClient, "team-match-b")

			test.DeleteHarborSyncConfig(k8sClient, "my-match-cfg")
		})

		It("should create secrets in namespace", func() {
			var err error
			mapping := crdv1.ProjectMapping{
				Type:      crdv1.MatchMappingType,
				Namespace: "team-match-.*",
				Secret:    "platform-pull-token",
			}
			cfg := test.EnsureHarborSyncConfigWithParams(k8sClient, "my-match-cfg", "platform-team", &mapping, nil)
			err = mapByMatching(
				k8sClient,
				mapping,
				cfg,
				harbor.Project{
					ID:   1,
					Name: "platform-team",
				},
				crdv1.RobotAccountCredential{
					Name:  "robot$sync-bot",
					Token: "my-token",
				},
				"my-registry-url",
			)
			Expect(err).ToNot(HaveOccurred())

			teamASecret := v1.Secret{}
			teamBSecret := v1.Secret{}
			SystemSecret := v1.Secret{}
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-match-a", Name: "platform-pull-token"}, &teamASecret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-match-b", Name: "platform-pull-token"}, &teamBSecret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "kube-system", Name: "platform-pull-token"}, &teamBSecret)
			Expect(err).To(HaveOccurred())

			Expect(string(teamASecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"my-registry-url":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(string(teamBSecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"my-registry-url":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(SystemSecret.Data).To(BeNil())
		})
	})

	Describe("Translate", func() {

		BeforeEach(func() {
			test.EnsureNamespace(k8sClient, "team-translate-a")
			test.EnsureNamespace(k8sClient, "team-translate-b")
		})

		AfterEach(func() {
			test.DeleteNamespace(k8sClient, "team-translate-a")
			test.DeleteNamespace(k8sClient, "team-translate-b")

			test.DeleteHarborSyncConfig(k8sClient, "my-translate-cfg")
		})

		It("should create secrets in namespace", func() {
			var err error
			mapping := crdv1.ProjectMapping{
				Type:      crdv1.MatchMappingType,
				Namespace: "team-translate-$1",
				Secret:    "team-$1-pull-token",
			}
			cfg := test.EnsureHarborSyncConfigWithParams(k8sClient, "my-translate-cfg", "team-translate-(.*)", &mapping, nil)

			mapByTranslating(
				k8sClient,
				mapping,
				cfg,
				harbor.Project{
					ID:   1,
					Name: "team-translate-a",
				},
				crdv1.RobotAccountCredential{
					Name:  "robot$sync-bot",
					Token: "my-token",
				},
				"my-registry-url",
			)

			teamASecret := v1.Secret{}
			teamBSecret := v1.Secret{}
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-translate-a", Name: "team-a-pull-token"}, &teamASecret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-translate-b", Name: "team-b-pull-token"}, &teamBSecret)
			Expect(err).To(HaveOccurred())

			Expect(string(teamASecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"my-registry-url":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(teamBSecret.Data).To(BeNil())
		})
	})

})

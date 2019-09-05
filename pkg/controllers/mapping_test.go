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
	"regexp"

	"github.com/go-logr/logr"
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

	var fakeHarbor harborfake.Client
	var log logr.Logger
	var hscr *HarborSyncConfigReconciler

	BeforeEach(func() {
		fakeHarbor = harborfake.Client{}
		log = zap.Logger(true)
		hscr = &HarborSyncConfigReconciler{
			k8sClient,
			log,
			fakeHarbor,
		}
	})

	Describe("Match", func() {

		BeforeEach(func() {
			ensureNamespace(k8sClient, "team-match-a")
			ensureNamespace(k8sClient, "team-match-b")
		})

		AfterEach(func() {
			deleteNamespace(k8sClient, "team-match-a")
			deleteNamespace(k8sClient, "team-match-b")

			deleteHarborSyncConfig(k8sClient, "my-match-cfg")
		})

		It("should create secrets in namespace", func(done Done) {
			var err error
			mapping := crdv1.ProjectMapping{
				Type:      crdv1.MatchMappingType,
				Namespace: "team-match-.*",
				Secret:    "platform-pull-token",
			}
			ensureHarborSyncConfigWithParams(k8sClient, "my-match-cfg", "platform-team", mapping, nil)
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
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-match-a", Name: "platform-pull-token"}, &teamASecret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-match-b", Name: "platform-pull-token"}, &teamBSecret)
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

		BeforeEach(func() {
			ensureNamespace(k8sClient, "team-translate-a")
			ensureNamespace(k8sClient, "team-translate-b")
		})

		AfterEach(func() {
			deleteNamespace(k8sClient, "team-translate-a")
			deleteNamespace(k8sClient, "team-translate-b")

			deleteHarborSyncConfig(k8sClient, "my-translate-cfg")
		})

		It("should create secrets in namespace", func(done Done) {
			var err error
			mapping := crdv1.ProjectMapping{
				Type:      crdv1.MatchMappingType,
				Namespace: "team-translate-$1",
				Secret:    "team-$1-pull-token",
			}
			ensureHarborSyncConfigWithParams(k8sClient, "my-translate-cfg", "team-translate-(.*)", mapping, nil)

			hscr.mapByTranslating(
				mapping,
				regexp.MustCompile("team-translate-(.*)"),
				harbor.Project{
					ID:   1,
					Name: "team-translate-a",
				},
				crdv1.RobotAccountCredential{
					Name:  "robot$sync-bot",
					Token: "my-token",
				},
			)

			teamASecret := v1.Secret{}
			teamBSecret := v1.Secret{}
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-translate-a", Name: "team-a-pull-token"}, &teamASecret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "team-translate-b", Name: "team-b-pull-token"}, &teamBSecret)
			Expect(err).To(HaveOccurred())

			Expect(string(teamASecret.Data[v1.DockerConfigJsonKey])).To(Equal(`{"auths":{"":{"username":"robot$sync-bot","password":"my-token","auth":"cm9ib3Qkc3luYy1ib3Q6bXktdG9rZW4="}}}`))
			Expect(teamBSecret.Data).To(BeNil())

			close(done)
		})
	})

})

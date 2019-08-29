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
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
	harborfake "github.com/moolen/harbor-sync/pkg/harbor/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Controller", func() {

	var harborClient harborfake.Client
	log := zap.Logger(false)

	BeforeEach(func() {
		harborClient = harborfake.Client{}
	})

	Describe("Robot", func() {

		It("should reconcile robot accounts", func(done Done) {
			mapping := crdv1.ProjectMapping{
				Namespace: "foo",
				Secret:    "foo",
				Type:      crdv1.TranslateMappingType,
			}
			cfg := ensureHarborSyncConfigWithParams(k8sClient, "my-cfg", "my-project", mapping)
			harborProject := harbor.Project{
				ID:   1,
				Name: "foo",
			}
			createdAccount := &harbor.CreateRobotResponse{
				Name:  "robot$sync-bot",
				Token: "1234",
			}
			harborClient.CreateRobotAccountFunc = func(name string, project harbor.Project) (*harbor.CreateRobotResponse, error) {
				return createdAccount, nil
			}
			statusCredentials := make(map[string]crdv1.RobotAccountCredentials)
			skip, credentials := reconcileRobotAccounts(harborClient, log, statusCredentials, harborProject, cfg.Spec.ProjectSelector[0].RobotAccountSuffix)
			Expect(skip).To(BeFalse())
			Expect(*credentials).To(Equal(crdv1.RobotAccountCredential{
				Name:  createdAccount.Name,
				Token: createdAccount.Token,
			}))
			Expect(statusCredentials["foo"]).To(Equal(crdv1.RobotAccountCredentials{
				crdv1.RobotAccountCredential{
					Name:  createdAccount.Name,
					Token: createdAccount.Token,
				},
			}))
			close(done)
		})
	})
})

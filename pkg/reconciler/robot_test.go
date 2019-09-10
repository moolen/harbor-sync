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
	"time"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
	harborfake "github.com/moolen/harbor-sync/pkg/harbor/fake"
	"github.com/moolen/harbor-sync/pkg/store"
	"github.com/moolen/harbor-sync/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Controller", func() {

	var harborClient harborfake.Client
	credStore, _ := store.NewTemp()
	mapping := crdv1.ProjectMapping{
		Namespace: "foo",
		Secret:    "foo",
		Type:      crdv1.TranslateMappingType,
	}
	harborProject := harbor.Project{
		ID:   1,
		Name: "foo",
	}
	createdAccount := &harbor.CreateRobotResponse{
		Name:  "robot$sync-bot",
		Token: "1234",
	}
	log := zap.Logger(false)

	BeforeEach(func() {
		harborClient = harborfake.Client{
			CreateRobotAccountFunc: func(name string, project harbor.Project) (*harbor.CreateRobotResponse, error) {
				return createdAccount, nil
			},
			GetRobotAccountsFunc: func(project harbor.Project) ([]harbor.Robot, error) {
				return []harbor.Robot{
					{
						Name:         "robot$sync-bot",
						CreationTime: "2222-01-02T15:04:05.999999999Z",
						ExpiresAt:    time.Now().UTC().Add(time.Hour * 24).Unix(),
					},
				}, nil
			},
		}
	})

	AfterEach(func() {
		err := credStore.Reset()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Robot", func() {

		It("should reconcile robot accounts", func(done Done) {
			cfg := test.EnsureHarborSyncConfigWithParams(k8sClient, "my-cfg", "my-project", &mapping, nil)
			harborClient.GetRobotAccountsFunc = nil
			credentials, changed, err := ReconcileRobotAccounts(
				harborClient,
				credStore,
				log,
				harborProject,
				cfg.Spec.RobotAccountSuffix,
				time.Hour*1,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			Expect(credentials.Name).To(Equal(createdAccount.Name))
			Expect(credentials.Token).To(Equal(createdAccount.Token))
			cacheCreds, err := credStore.Get("foo", "robot$sync-bot")
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheCreds.Name).To(Equal(createdAccount.Name))
			Expect(cacheCreds.Token).To(Equal(createdAccount.Token))
			close(done)
		})

		It("should use robot credentials from store", func(done Done) {
			cfg := test.EnsureHarborSyncConfigWithParams(k8sClient, "my-cfg", "my-project", &mapping, nil)
			harborClient.CreateRobotAccountFunc = nil
			credStore.Set("foo", crdv1.RobotAccountCredential{Name: "robot$sync-bot", Token: "bar"})
			credentials, changed, err := ReconcileRobotAccounts(
				harborClient,
				credStore,
				log,
				harborProject,
				cfg.Spec.RobotAccountSuffix,
				time.Hour*1,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeFalse())
			Expect(credentials.Name).To(Equal("robot$sync-bot"))
			Expect(credentials.Token).To(Equal("bar"))
			cacheCreds, err := credStore.Get("foo", "robot$sync-bot")
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheCreds.Name).To(Equal("robot$sync-bot"))
			Expect(cacheCreds.Token).To(Equal("bar"))
			close(done)
		})

		It("should delete robot account when credentials are missing", func(done Done) {
			cfg := test.EnsureHarborSyncConfigWithParams(k8sClient, "my-cfg", "my-project", &mapping, nil)
			var deleteCalled bool
			harborClient.DeleteRobotAccountFunc = func(project harbor.Project, robotID int) error {
				deleteCalled = true
				return nil
			}
			credentials, changed, err := ReconcileRobotAccounts(
				harborClient,
				credStore,
				log,
				harborProject,
				cfg.Spec.RobotAccountSuffix,
				time.Hour*1,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			Expect(deleteCalled).To(BeTrue())
			Expect(credentials.Name).To(Equal(createdAccount.Name))
			Expect(credentials.Token).To(Equal(createdAccount.Token))
			cacheCreds, err := credStore.Get("foo", "robot$sync-bot")
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheCreds.Name).To(Equal(createdAccount.Name))
			Expect(cacheCreds.Token).To(Equal(createdAccount.Token))
			close(done)
		})

		It("should re-create disabled account", func(done Done) {
			cfg := test.EnsureHarborSyncConfigWithParams(k8sClient, "my-cfg", "my-project", &mapping, nil)
			harborClient.GetRobotAccountsFunc = func(project harbor.Project) ([]harbor.Robot, error) {
				return []harbor.Robot{
					{
						Name:         "robot$sync-bot",
						Disabled:     true,
						CreationTime: "2222-01-02T15:04:05.999999999Z",
						ExpiresAt:    time.Now().UTC().Add(time.Hour * 24).Unix(),
					},
				}, nil
			}
			var deleteCalled bool
			harborClient.DeleteRobotAccountFunc = func(project harbor.Project, robotID int) error {
				deleteCalled = true
				return nil
			}
			credStore.Set("foo", crdv1.RobotAccountCredential{Name: "robot$sync-bot", Token: "bar"})
			credentials, changed, err := ReconcileRobotAccounts(
				harborClient,
				credStore,
				log,
				harborProject,
				cfg.Spec.RobotAccountSuffix,
				time.Hour*1,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			Expect(deleteCalled).To(BeTrue())
			Expect(credentials.Name).To(Equal(createdAccount.Name))
			Expect(credentials.Token).To(Equal(createdAccount.Token))
			cacheCreds, err := credStore.Get("foo", "robot$sync-bot")
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheCreds.Name).To(Equal(createdAccount.Name))
			Expect(cacheCreds.Token).To(Equal(createdAccount.Token))
			close(done)
		})

		It("should re-create expiring account", func(done Done) {
			cfg := test.EnsureHarborSyncConfigWithParams(k8sClient, "my-cfg", "my-project", &mapping, nil)
			harborClient.GetRobotAccountsFunc = func(project harbor.Project) ([]harbor.Robot, error) {
				return []harbor.Robot{
					{
						Name:         "robot$sync-bot",
						CreationTime: "2222-01-02T15:04:05.999999999Z",
						ExpiresAt:    0,
					},
				}, nil
			}
			var deleteCalled bool
			harborClient.DeleteRobotAccountFunc = func(project harbor.Project, robotID int) error {
				deleteCalled = true
				return nil
			}
			credStore.Set("foo", crdv1.RobotAccountCredential{Name: "robot$sync-bot", Token: "bar"})
			credentials, changed, err := ReconcileRobotAccounts(
				harborClient,
				credStore,
				log,
				harborProject,
				cfg.Spec.RobotAccountSuffix,
				time.Hour*1,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			Expect(deleteCalled).To(BeTrue())
			Expect(credentials.Name).To(Equal(createdAccount.Name))
			Expect(credentials.Token).To(Equal(createdAccount.Token))
			cacheCreds, err := credStore.Get("foo", "robot$sync-bot")
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheCreds.Name).To(Equal(createdAccount.Name))
			Expect(cacheCreds.Token).To(Equal(createdAccount.Token))
			close(done)
		})

		It("should rotate account", func(done Done) {
			cfg := test.EnsureHarborSyncConfigWithParams(k8sClient, "my-cfg", "my-project", &mapping, nil)
			harborClient.GetRobotAccountsFunc = func(project harbor.Project) ([]harbor.Robot, error) {
				return []harbor.Robot{
					{
						Name:         "robot$sync-bot",
						CreationTime: time.Now().UTC().Add(-time.Hour * 2).Format(time.RFC3339Nano),
						ExpiresAt:    time.Now().UTC().Add(time.Hour * 24).Unix(),
					},
				}, nil
			}
			var deleteCalled bool
			harborClient.DeleteRobotAccountFunc = func(project harbor.Project, robotID int) error {
				deleteCalled = true
				return nil
			}
			credStore.Set("foo", crdv1.RobotAccountCredential{Name: "robot$sync-bot", Token: "bar"})
			credentials, changed, err := ReconcileRobotAccounts(
				harborClient,
				credStore,
				log,
				harborProject,
				cfg.Spec.RobotAccountSuffix,
				time.Hour*1,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			Expect(deleteCalled).To(BeTrue())
			Expect(credentials.Name).To(Equal(createdAccount.Name))
			Expect(credentials.Token).To(Equal(createdAccount.Token))
			cacheCreds, err := credStore.Get("foo", "robot$sync-bot")
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheCreds.Name).To(Equal(createdAccount.Name))
			Expect(cacheCreds.Token).To(Equal(createdAccount.Token))
			close(done)
		})
	})
})

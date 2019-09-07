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

package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/moolen/harbor-sync/pkg/harbor"
	"github.com/moolen/harbor-sync/pkg/harbor/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var errListProjects = fmt.Errorf("error listing projects")
var errListRobots = fmt.Errorf("error listing robots")

var _ = Describe("Repository", func() {
	It("should cache the state of projects and robot accounts", func() {
		c := &fake.Client{}
		log := zap.Logger(false)
		rep, err := New(c, log, time.Second*500)
		Expect(err).ToNot(HaveOccurred())
		c.BaseURLFunc = func() string { return "myurl" }
		c.ListProjectsFunc = func() ([]harbor.Project, error) {
			return []harbor.Project{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			}, nil
		}
		c.GetRobotAccountsFunc = func(p harbor.Project) ([]harbor.Robot, error) {
			if p.Name == "foo" {
				return []harbor.Robot{
					{
						Name: "foo-1",
					},
					{
						Name: "foo-2",
					},
				}, nil
			}
			return []harbor.Robot{
				{
					Name: "bar-1",
				},
				{
					Name: "bar-2",
				},
			}, nil
		}
		Expect(rep.StateHash).To(Equal(uint64(0)))
		err = rep.Update()
		Expect(err).ToNot(HaveOccurred())

		// expect to that cache has been set
		// and state was updated
		Expect(rep.StateHash).ToNot(Equal(uint64(0)))

		rr, err := rep.GetRobotAccounts(harbor.Project{Name: "foo"})
		Expect(err).ToNot(HaveOccurred())
		Expect(rr).To(HaveLen(2))
		rr, err = rep.GetRobotAccounts(harbor.Project{Name: "foo"})
		Expect(err).ToNot(HaveOccurred())
		Expect(rr).To(HaveLen(2))

		Expect(rep.ProjectsCache.Get()).To(HaveLen(2))
		Expect(rep.RobotsCache.Get("foo")).To(HaveLen(2))
		Expect(rep.RobotsCache.Get("bar")).To(HaveLen(2))
		Expect(rep.ListProjects()).To(HaveLen(2))
		Expect(rep.BaseURL()).To(Equal("myurl"))

		// if project or robot calls fail:
		c.GetRobotAccountsFunc = func(p harbor.Project) ([]harbor.Robot, error) {
			return nil, errListRobots
		}
		Expect(rep.Update()).To(HaveOccurred())
		c.ListProjectsFunc = func() ([]harbor.Project, error) {
			return nil, errListProjects
		}
		Expect(rep.Update()).To(HaveOccurred())
	})

	It("should sync with the API", func() {
		var projects []harbor.Project
		var robots []harbor.Robot
		c := &fake.Client{}
		log := zap.Logger(false)
		rep, err := New(c, log, time.Millisecond*10)
		Expect(err).ToNot(HaveOccurred())
		c.ListProjectsFunc = func() ([]harbor.Project, error) {
			return projects, nil
		}
		c.GetRobotAccountsFunc = func(p harbor.Project) ([]harbor.Robot, error) {
			return robots, nil
		}

		// first: prep state
		err = rep.Update()
		Expect(err).ToNot(HaveOccurred())

		changeChan := rep.Sync()

		// change the repo state
		go func() {
			<-time.After(time.Millisecond * 20)
			projects = append(projects, harbor.Project{Name: "foo"})
		}()

		<-changeChan
	})
})

func TestRepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Repository Suite")
}

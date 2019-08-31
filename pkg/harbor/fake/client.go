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

package fake

import (
	"github.com/moolen/harbor-sync/pkg/harbor"
)

// Client implements harbor.API
// and offers a Func interface for each method
type Client struct {
	BaseURLFunc            func() string
	ListProjectsFunc       func() ([]harbor.Project, error)
	GetRobotAccountsFunc   func(project harbor.Project) ([]harbor.Robot, error)
	CreateRobotAccountFunc func(name string, project harbor.Project) (*harbor.CreateRobotResponse, error)
	DeleteRobotAccountFunc func(project harbor.Project, robotID int) error
}

// ListProjects ...
func (f Client) ListProjects() ([]harbor.Project, error) {
	if f.ListProjectsFunc != nil {
		return f.ListProjectsFunc()
	}
	return []harbor.Project{}, nil
}

// GetRobotAccounts ...
func (f Client) GetRobotAccounts(project harbor.Project) ([]harbor.Robot, error) {
	if f.GetRobotAccountsFunc != nil {
		return f.GetRobotAccountsFunc(project)
	}
	return []harbor.Robot{}, nil
}

// CreateRobotAccount ...
func (f Client) CreateRobotAccount(name string, project harbor.Project) (*harbor.CreateRobotResponse, error) {
	if f.CreateRobotAccountFunc != nil {
		return f.CreateRobotAccountFunc(name, project)
	}
	return &harbor.CreateRobotResponse{}, nil
}

// DeleteRobotAccount ...
func (f Client) DeleteRobotAccount(project harbor.Project, robotID int) error {
	if f.DeleteRobotAccountFunc != nil {
		return f.DeleteRobotAccountFunc(project, robotID)
	}
	return nil
}

// BaseURL ...
func (f Client) BaseURL() string {
	if f.BaseURLFunc != nil {
		return f.BaseURLFunc()
	}
	return ""
}

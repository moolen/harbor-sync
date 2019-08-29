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
	"github.com/moolen/harbor-sync/pkg/harbor"
)

// GetRobotAccounts returns a list of robot accounts for the given project
func (r *Repository) GetRobotAccounts(project harbor.Project) ([]harbor.Robot, error) {
	return r.RobotsCache.Get(project.Name), nil
}

// CreateRobotAccount creates a robot account and updates the internal cache
func (r *Repository) CreateRobotAccount(name string, project harbor.Project) (*harbor.CreateRobotResponse, error) {
	res, err := r.Client.CreateRobotAccount(name, project)
	if err != nil {
		return res, err
	}
	accs, err := r.Client.GetRobotAccounts(project)
	if err != nil {
		return res, err
	}
	r.RobotsCache.Set(project.Name, accs)
	err = r.UpdateHash()
	if err != nil {
		return res, err
	}
	return res, nil
}

// DeleteRobotAccount deletes the robot account and updates the internal cache
func (r *Repository) DeleteRobotAccount(project harbor.Project, robotID int) error {
	err := r.Client.DeleteRobotAccount(project, robotID)
	if err != nil {
		return err
	}
	accs, err := r.Client.GetRobotAccounts(project)
	if err != nil {
		return err
	}
	r.RobotsCache.Set(project.Name, accs)
	err = r.UpdateHash()
	if err != nil {
		return err
	}
	return nil
}

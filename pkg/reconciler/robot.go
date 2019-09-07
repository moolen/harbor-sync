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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
)

const robotPrefix = "robot$"

// ReconcileRobotAccounts ensures that the required robot accounts exist in the given project
func ReconcileRobotAccounts(harborAPI harbor.API, log logr.Logger, syncConfig *crdv1.HarborSync, project harbor.Project, accountSuffix string) (*crdv1.RobotAccountCredential, bool, error) {
	robots, err := harborAPI.GetRobotAccounts(project)
	if err != nil {
		return nil, false, fmt.Errorf("could not get robot accounts from harbor")
	}

	// check if we manage the credentials for this robot account
	// if we do not have them we first delete, then re-create the robot account
	for _, robot := range robots {

		// only one robot account will match
		if robot.Name == addPrefix(accountSuffix) {
			log.V(1).Info("robot account already exists", "project_name", project.Name, "robot_account", robot.Name)
			fmt.Printf("creds: %#v", syncConfig.Status.RobotCredentials)
			// case: robot account exists in harbor, but we do not have the credentials: re-create!
			if syncConfig.Status.RobotCredentials == nil {
				log.Info(fmt.Sprintf("sync config status.credentials does not exist, deleting robot account"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					return nil, false, fmt.Errorf("could not delete robot account: %s", err.Error())
				}
			}

			// case: robot is disabled: re-create
			if robot.Disabled == true {
				log.Info(fmt.Sprintf("robot account is disabled, deleting it"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					return nil, false, fmt.Errorf("could not delete robot account: %s", err.Error())
				}
			}

			// case: robot will expires soon: re-create
			// TODO: implement token regeneration API once it is upstream available:
			// https://github.com/goharbor/harbor/issues/8405
			if expiresSoon(robot) {
				log.Info(fmt.Sprintf("robot account expires soon, deleting it"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					return nil, false, fmt.Errorf("could not delete robot account: %s", err.Error())
				}
			}

			// good case: we have the credentials. do not re-create
			cred := syncConfig.Status.RobotCredentials[project.Name]
			if cred.Name == addPrefix(accountSuffix) {
				log.V(1).Info("found credentials in status.credentials. will not delete robot account")
				return &cred, false, nil
			}

			// case: creds do not exist. delete robot account
			log.Info("sync config does not hold the credentials for robot account. deleting it.", "project_name", project.Name, "robot_account", robot.Name)
			err = harborAPI.DeleteRobotAccount(project, robot.ID)
			if err != nil {
				return nil, false, fmt.Errorf("could not delete robot account: %s", err)
			}
		}
	}

	log.Info("creating robot account", "project_name", project.Name, "robot_account_suffix", accountSuffix)
	res, err := harborAPI.CreateRobotAccount(accountSuffix, project)
	if err != nil {
		return nil, false, fmt.Errorf("could not create robot account")
	}
	// store secret in status field
	if syncConfig.Status.RobotCredentials == nil {
		syncConfig.Status.RobotCredentials = make(map[string]crdv1.RobotAccountCredential)
	}
	log.V(1).Info("updating status field", "project_name", project.Name)

	// check if old token exists: update it or append it to list
	cred := crdv1.RobotAccountCredential{
		Name:  res.Name,
		Token: res.Token,
	}
	syncConfig.Status.RobotCredentials[project.Name] = cred
	return &cred, true, nil
}

func addPrefix(str string) string {
	return robotPrefix + str
}

func expiresSoon(robot harbor.Robot) bool {
	now := time.Now().Add(time.Hour) // TODO: make this configurable
	expiry := time.Unix(int64(robot.ExpiresAt), 0)
	return expiry.Before(now)
}
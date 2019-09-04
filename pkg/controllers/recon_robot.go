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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/moolen/harbor-sync/pkg/harbor"
)

const robotPrefix = "robot$"

func reconcileRobotAccounts(harborAPI harbor.API, log logr.Logger, syncConfig *crdv1.HarborSync, project harbor.Project, accountSuffix string) (bool, *crdv1.RobotAccountCredential) {
	robots, err := harborAPI.GetRobotAccounts(project)
	if err != nil {
		log.Error(err, "could not get robot accounts from harbor")
		return true, nil
	}

	// check if we manage the credentials for this robot account
	// if we do not have them we first delete, then re-create the robot account
	for _, robot := range robots {
		if robot.Name == addPrefix(accountSuffix) {
			log.V(1).Info("robot account already exists", "project_name", project.Name, "robot_account", robot.Name)

			// case: robot account exists in harbor, but we do not have the credentials: re-create!
			if syncConfig.Status.RobotCredentials == nil {
				log.Info(fmt.Sprintf("sync config status.credentials does not exist, deleting robot account"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					log.Error(err, "could not delete robot account", "project_name", project.Name, "robot_account", robot.Name)
				}
				continue
			}

			// case: robot is disabled: re-create
			if robot.Disabled == true {
				log.Info(fmt.Sprintf("robot account is disabled, deleting it"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					log.Error(err, "could not delete robot account", "project_name", project.Name, "robot_account", robot.Name)
				}
				continue
			}

			// case: robot will expires soon: re-create
			// TODO: implement token regeneration API once it is upstream available:
			// https://github.com/goharbor/harbor/issues/8405
			if expiresSoon(robot) {
				log.Info(fmt.Sprintf("robot account expires soon, deleting it"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					log.Error(err, "could not delete robot account", "project_name", project.Name, "robot_account", robot.Name)
				}
				continue
			}

			// good case: we have the credentials. do not re-create
			creds := syncConfig.Status.RobotCredentials[project.Name]
			for _, cred := range creds {
				if cred.Name == addPrefix(accountSuffix) {
					log.V(1).Info("found credentials in status.credentials. will not delete robot account")
					return false, &cred
				}
			}

			// case: creds do not exist. delete robot account
			log.Info("sync config does not hold the credentials for robot account. deleting it.", "project_name", project.Name, "robot_account", robot.Name)
			err = harborAPI.DeleteRobotAccount(project, robot.ID)
			if err != nil {
				log.Error(err, "could not delete robot account", "project_name", project.Name, "robot_account", robot.Name)
				continue
			}
		}
	}

	log.Info("creating robot account", "project_name", project.Name, "robot_account_suffix", accountSuffix)
	res, err := harborAPI.CreateRobotAccount(accountSuffix, project)
	if err != nil {
		log.Error(err, "could not create robot account", "project_name", project.Name)
		return true, nil
	}
	// store secret in status field
	if syncConfig.Status.RobotCredentials == nil {
		syncConfig.Status.RobotCredentials = make(map[string]crdv1.RobotAccountCredentials)
	}
	if syncConfig.Status.RobotCredentials[project.Name] == nil {
		syncConfig.Status.RobotCredentials[project.Name] = crdv1.RobotAccountCredentials{}
	}
	log.Info("updating status field", "project_name", project.Name)

	// check if old token exists: update it or append it to list
	found := false
	var credential crdv1.RobotAccountCredential
	creds := syncConfig.Status.RobotCredentials[project.Name]
	for i, cred := range creds {
		if cred.Name == addPrefix(accountSuffix) {
			log.V(1).Info("found credentials in status.credentials. updating token")
			creds[i].Token = res.Token
			found = true
			credential = creds[i]
			break
		}
	}
	if !found {
		credential = crdv1.RobotAccountCredential{
			Name:  res.Name,
			Token: res.Token,
		}
		creds = append(creds, credential)
	}

	syncConfig.Status.RobotCredentials[project.Name] = creds
	return false, &credential
}

func addPrefix(str string) string {
	return robotPrefix + str
}

func expiresSoon(robot harbor.Robot) bool {
	now := time.Now().Add(time.Hour) // TODO: make this configurable
	expiry := time.Unix(int64(robot.ExpiresAt), 0)
	return expiry.Before(now)
}

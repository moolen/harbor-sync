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

// CredentialStore is an interface that is used to store the credentials
type CredentialStore interface {
	Has(project, name string) bool
	Get(project, name string) (*crdv1.RobotAccountCredential, error)
	Set(project string, cred crdv1.RobotAccountCredential) error
	Reset() error
}

// ReconcileRobotAccounts ensures that the required robot accounts exist in the given project
func ReconcileRobotAccounts(
	harborAPI harbor.API,
	creds CredentialStore,
	log logr.Logger,
	project harbor.Project,
	accountSuffix string,
	rotationInterval time.Duration,
) (*crdv1.RobotAccountCredential, bool, error) {
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
			haveCredentials := creds.Has(project.Name, addPrefix(accountSuffix))
			existingCreds, _ := creds.Get(project.Name, addPrefix(accountSuffix))

			// case: robot account exists in harbor, but we do not have the credentials: re-create!
			if !haveCredentials {
				log.Info(fmt.Sprintf("store does not have credentials, deleting robot account"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					return nil, false, fmt.Errorf("could not delete robot account: %s", err.Error())
				}
				break
			}

			// case: robot is disabled: re-create
			if robot.Disabled == true {
				log.Info(fmt.Sprintf("robot account is disabled, deleting it"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					return nil, false, fmt.Errorf("could not delete robot account: %s", err.Error())
				}
				break
			}

			if shouldRotate(robot, rotationInterval) {
				log.Info(fmt.Sprintf("robot account should rotate, deleting it"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					return nil, false, fmt.Errorf("could not delete robot account: %s", err.Error())
				}
				break
			}

			// case: robot will expires soon: re-create
			// TODO: implement token regeneration API once it is upstream available:
			// https://github.com/goharbor/harbor/issues/8405
			if expiresSoon(robot, rotationInterval) {
				log.Info(fmt.Sprintf("robot account expires soon, deleting it"))
				err = harborAPI.DeleteRobotAccount(project, robot.ID)
				if err != nil {
					return nil, false, fmt.Errorf("could not delete robot account: %s", err.Error())
				}
				break
			}

			// good case: we have the credentials. do not re-create
			log.V(1).Info("found credentials in store. will not delete robot account")
			return existingCreds, false, nil
		}
	}

	log.Info("creating robot account", "project_name", project.Name, "robot_account_suffix", accountSuffix)
	res, err := harborAPI.CreateRobotAccount(accountSuffix, project)
	if err != nil {
		return nil, false, fmt.Errorf("could not create robot account")
	}

	// check if old token exists: update it or append it to list
	cred := crdv1.RobotAccountCredential{
		Name:      res.Name,
		CreatedAt: time.Now().UTC().Unix(),
		Token:     res.Token,
	}

	log.V(1).Info("updating store with credentials", "project_name", project.Name)
	err = creds.Set(project.Name, cred)
	if err != nil {
		return nil, true, err
	}
	return &cred, true, nil
}

func addPrefix(str string) string {
	return robotPrefix + str
}

func shouldRotate(robot harbor.Robot, interval time.Duration) bool {
	created, err := time.Parse(time.RFC3339Nano, robot.CreationTime)
	if err != nil {
		fmt.Printf("error parsing time: %s\n", err.Error())
		return true
	}
	return created.UTC().Add(interval).Before(time.Now().UTC())
}

func expiresSoon(robot harbor.Robot, duration time.Duration) bool {
	now := time.Now().UTC().Add(duration)
	expiry := time.Unix(robot.ExpiresAt, 0)
	return expiry.Before(now)
}

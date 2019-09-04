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

package harbor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Robot is the API response from Harbor
type Robot struct {
	ID        int `json:"id"`
	ProjectID int `json:"project_id"`
	Disabled  bool

	Name         string `json:"name"`
	Description  string `json:"description"`
	ExpiresAt    int    `json:"expires_at"`
	CreationTime string `json:"creation_time"`
	UpdateTime   string `json:"update_time"`
}

// CreateRobotRequest is the request payload for creating a robot account
type CreateRobotRequest struct {
	Name   string                     `json:"name"`
	Access []CreateRobotRequestAccess `json:"access"`
}

// CreateRobotRequestAccess defines the permissions for the robot account
type CreateRobotRequestAccess struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// CreateRobotResponse is the API response from a creating a robot
type CreateRobotResponse struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

// GetRobotAccounts returns all robot accounts for the given project
func (c *Client) GetRobotAccounts(project Project) ([]Robot, error) {
	var robotAccounts []Robot
	resp, err := c.newRequest("GET", fmt.Sprintf("/api/projects/%d/robots", project.ID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&robotAccounts)
	if err != nil {
		return robotAccounts, err
	}

	for _, acc := range robotAccounts {
		robotAccountExpiry.WithLabelValues(project.Name, acc.Name).Set(float64(acc.ExpiresAt))
	}

	return robotAccounts, nil
}

// CreateRobotAccount creates a robot account and return the name and token
func (c *Client) CreateRobotAccount(name string, project Project) (*CreateRobotResponse, error) {
	var robotResponse CreateRobotResponse
	reqBody, err := json.Marshal(CreateRobotRequest{
		Name: name,
		Access: []CreateRobotRequestAccess{
			{
				Resource: fmt.Sprintf("/project/%d/repository", project.ID),
				Action:   "pull",
			},
			{
				Resource: fmt.Sprintf("/project/%s/repository", project.Name),
				Action:   "pull",
			},
		},
	})

	if err != nil {
		return nil, err
	}
	resp, err := c.newRequest("POST", fmt.Sprintf("/api/projects/%d/robots", project.ID), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("could not create new http request: %s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %s", err.Error())
	}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&robotResponse)
	if err != nil {
		return nil, fmt.Errorf("could not decode response body: %s", err.Error())
	}

	return &robotResponse, nil
}

// DeleteRobotAccount deletes the specified robot account
func (c *Client) DeleteRobotAccount(project Project, robotID int) error {
	resp, err := c.newRequest("DELETE", fmt.Sprintf("/api/projects/%d/robots/%d", project.ID, robotID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

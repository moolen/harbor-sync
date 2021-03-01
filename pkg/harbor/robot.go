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
	"strings"

	"github.com/blang/semver"
)

// Robot is the API response from Harbor
type Robot struct {
	ID        int `json:"id"`
	ProjectID int `json:"project_id"`
	Disabled  bool

	Name         string `json:"name"`
	Description  string `json:"description"`
	ExpiresAt    int64  `json:"expires_at"`
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

// CreateRobotResponseV22 is the API response from a creating a robot
type CreateRobotResponseV22 struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

// GetRobotAccounts returns all robot accounts for the given project
func (c *Client) GetRobotAccounts(project Project) ([]Robot, error) {
	var robotAccounts []Robot
	resp, err := c.newRequest("GET", fmt.Sprintf("projects/%d/robots", project.ID), nil)
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

	// remove labels for non-existent robots
	c.mu.Lock()
	if last, ok := c.lastRobotAccounts[project.Name]; ok {
		diff := diffRobotAccounts(last, robotAccounts)
		for _, rname := range diff {
			robotAccountExpiry.DeleteLabelValues(project.Name, rname)
		}
	}
	c.lastRobotAccounts[project.Name] = toRobotNames(robotAccounts)
	c.mu.Unlock()

	return robotAccounts, nil
}

// CreateRobotAccount creates a robot account and return the name and token
func (c *Client) CreateRobotAccount(name string, pushAccess bool, project Project) (*CreateRobotResponse, error) {
	// we have to check the api version before we make an API call
	// there is a quirk with the harbor API w/ permissions
	info, err := c.SystemInfo()
	if err != nil {
		return nil, fmt.Errorf("error calling system info: %s", err)
	}
	v, err := semver.Make(strings.TrimLeft(info.HarborVersion, "v"))
	if err != nil {
		return nil, fmt.Errorf("unable to parse harbor version")
	}
	v.Pre = nil
	v110 := semver.MustParse("1.10.0")
	permissions := []CreateRobotRequestAccess{
		{
			Resource: fmt.Sprintf("/project/%d/repository", project.ID),
			Action:   "pull",
		},
	}

	if v.LT(v110) {
		permissions = append(permissions, CreateRobotRequestAccess{
			Resource: fmt.Sprintf("/project/%s/repository", project.Name),
			Action:   "pull",
		})
	}

	if pushAccess {
		permissions = append(permissions, CreateRobotRequestAccess{
			Resource: fmt.Sprintf("/project/%d/repository", project.ID),
			Action:   "push",
		})
		if v.LT(v110) {
			permissions = append(permissions, CreateRobotRequestAccess{
				Resource: fmt.Sprintf("/project/%s/repository", project.Name),
				Action:   "push",
			})
		}
	}

	reqBody, err := json.Marshal(CreateRobotRequest{
		Name:   name,
		Access: permissions,
	})

	if err != nil {
		return nil, err
	}
	resp, err := c.newRequest("POST", fmt.Sprintf("projects/%d/robots", project.ID), bytes.NewReader(reqBody))
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

	// robot api has changed in v2.2.0
	var robotResponse CreateRobotResponse
	v220 := semver.MustParse("2.2.0")
	if v.GTE(v220) {
		var r22 CreateRobotResponseV22
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&r22)
		if err != nil {
			return nil, fmt.Errorf("could not decode response body: %s", err.Error())
		}
		return &CreateRobotResponse{
			Name:  r22.Name,
			Token: r22.Secret,
		}, nil
	}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&robotResponse)
	if err != nil {
		return nil, fmt.Errorf("could not decode response body: %s", err.Error())
	}

	return &robotResponse, nil
}

// DeleteRobotAccount deletes the specified robot account
func (c *Client) DeleteRobotAccount(project Project, robotID int) error {
	resp, err := c.newRequest("DELETE", fmt.Sprintf("projects/%d/robots/%d", project.ID, robotID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// diffRobotAccounts returns the difference between last and currently observed robot accounts
func diffRobotAccounts(last []string, current []Robot) []string {
	var diff []string
	c := toMap(current)
	for _, l := range last {
		if !c[l] {
			// old entry is missing now
			diff = append(diff, l)
		}
	}
	return diff
}

// toMap transforms a robot list to a map
func toMap(rr []Robot) map[string]bool {
	m := make(map[string]bool)
	for _, r := range rr {
		m[r.Name] = true
	}
	return m
}

// toRobotNames returns a list of robot account names
func toRobotNames(rr []Robot) []string {
	var out []string
	for _, r := range rr {
		out = append(out, r.Name)
	}
	return out
}

package harbor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

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

type CreateRobotRequest struct {
	Name   string                     `json:"name"`
	Access []CreateRobotRequestAccess `json:"access"`
}
type CreateRobotResponse struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type CreateRobotRequestAccess struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

func (c *Client) GetRobotAccounts(project Project) ([]Robot, error) {
	var robotAccounts []Robot
	robotsURL, err := url.ParseRequestURI(fmt.Sprintf("/api/projects/%d/robots", project.ID))
	if err != nil {
		return nil, err
	}
	u := c.APIBaseURL.ResolveReference(robotsURL)
	req, err := c.newRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
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

	return robotAccounts, nil
}

func (c *Client) CreateRobotAccount(name string, project Project) (*CreateRobotResponse, error) {
	var robotResponse CreateRobotResponse
	robotsURL, err := url.ParseRequestURI(fmt.Sprintf("/api/projects/%d/robots", project.ID))
	if err != nil {
		return nil, err
	}
	u := c.APIBaseURL.ResolveReference(robotsURL)

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
	req, err := c.newRequest("POST", u.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("could not create new http request: %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not issue http request: %s", err.Error())
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

func (c *Client) DeleteRobotAccount(project Project, robotID int) error {
	robotsURL, err := url.ParseRequestURI(fmt.Sprintf("/api/projects/%d/robots/%d", project.ID, robotID))
	if err != nil {
		return err
	}
	u := c.APIBaseURL.ResolveReference(robotsURL)
	req, err := c.newRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	}

	return fmt.Errorf("unexpected status: %s", resp.Status)
}

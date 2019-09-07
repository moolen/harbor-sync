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

// API is the interface that is implemented by the client
// and repository type
type API interface {
	ListProjects() ([]Project, error)
	GetRobotAccounts(project Project) ([]Robot, error)
	CreateRobotAccount(name string, project Project) (*CreateRobotResponse, error)
	DeleteRobotAccount(project Project, robotID int) error
	BaseURL() string
}

// Project is the harbor API response
type Project struct {
	ID        int             `json:"project_id"`
	Name      string          `json:"name"`
	OwnerName string          `json:"owner_name"`
	Metadata  ProjectMetadata `json:"metadata"`
}

// ProjectMetadata contains the metadata of the project
type ProjectMetadata struct {
	Public             string `json:"public"`
	EnableContentTrust string `json:"enable_content_trust"`
	PreventVul         string `json:"prevent_vul"`
	Severity           string `json:"severity"`
	AutoScan           string `json:"auto_scal"`
}

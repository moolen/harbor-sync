package fake

import (
	"github.com/moolen/harbor-sync/pkg/harbor"
)

// Client implementes harbor.API
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

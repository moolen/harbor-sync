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

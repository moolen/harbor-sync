package repository

import (
	"github.com/moolen/harbor-sync/pkg/harbor"
)

// ListProjects implements the harbor.API
// it returns all harbor projects
func (r *Repository) ListProjects() ([]harbor.Project, error) {
	return r.ProjectsCache.Get(), nil
}

package repository

import (
	"sync"

	"github.com/moolen/harbor-sync/pkg/harbor"
)

type ProjectsCache struct {
	mu   *sync.RWMutex
	data map[string]harbor.Project
}

type RobotsCache struct {
	mu   *sync.RWMutex
	data map[string][]harbor.Robot
}

func (p *ProjectsCache) Set(project harbor.Project) {
	p.mu.Lock()
	p.data[project.Name] = project
	p.mu.Unlock()
}

func (p *ProjectsCache) Get() []harbor.Project {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var projects []harbor.Project
	for _, project := range p.data {
		projects = append(projects, project)
	}
	return projects
}

func (r *RobotsCache) Set(key string, val []harbor.Robot) {
	r.mu.Lock()
	r.data[key] = val
	r.mu.Unlock()
}

func (r *RobotsCache) Get(key string) []harbor.Robot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.data[key]
}

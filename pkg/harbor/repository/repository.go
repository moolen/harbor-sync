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

package repository

import (
	"sync"
	"time"

	Logr "github.com/go-logr/logr"
	"github.com/mitchellh/hashstructure"
	"github.com/moolen/harbor-sync/pkg/harbor"
)

// Repository implements the harbor.API interface
// it caches the API resources using a polling mechanism in the background
type Repository struct {
	Log          Logr.Logger
	Client       harbor.API
	PollInterval time.Duration

	// StateHash is a uniqe number which is computed
	// with the projects and robots structs.
	// This is used to compare the old/new version of data
	StateHash     uint64
	ProjectsCache *ProjectsCache
	RobotsCache   *RobotsCache
}

// New is the repository constructor
func New(client harbor.API, logger Logr.Logger, interval time.Duration) (*Repository, error) {
	return &Repository{
		Log:          logger,
		Client:       client,
		PollInterval: interval,
		ProjectsCache: &ProjectsCache{
			mu:   &sync.RWMutex{},
			data: make(map[string]harbor.Project),
		},
		RobotsCache: &RobotsCache{
			mu:   &sync.RWMutex{},
			data: make(map[string][]harbor.Robot),
		},
	}, nil
}

func (r *Repository) BaseURL() string {
	return r.Client.BaseURL()
}

// Update fetches the projects and robot accounts
func (r *Repository) Update() error {
	var err error
	projects, err := r.Client.ListProjects()
	if err != nil {
		r.Log.Error(err, "error listing projects")
		return err
	}
	r.Log.V(1).Info("listing projects", "found_projects", len(projects))

	for _, project := range projects {
		r.ProjectsCache.Set(project)
		robotAccounts, err := r.Client.GetRobotAccounts(project)
		if err != nil {
			r.Log.Error(err, "error fetching robot account", "project_name", project.Name)
			return err
		}
		r.Log.V(1).Info("listing robot accounts", "found_robot_accouns", len(robotAccounts), "project_name", project.Name)
		r.RobotsCache.Set(project.Name, robotAccounts)

	}
	r.UpdateHash()
	return nil
}

// UpdateHash recalculates the StateHash
func (r *Repository) UpdateHash() error {
	var robotsHash uint64
	projects := r.ProjectsCache.Get()

	for _, project := range projects {
		robotAccounts := r.RobotsCache.Get(project.Name)
		rh, err := hashstructure.Hash(robotAccounts, nil)
		if err != nil {
			r.Log.Error(err, "could not hash robot accounts")
			return err
		}
		robotsHash += rh
	}

	projectsHash, err := hashstructure.Hash(projects, nil)
	if err != nil {
		r.Log.Error(err, "could not hash projects")
		return err
	}
	r.StateHash = projectsHash + robotsHash
	return nil
}

// Sync returns a channel which notifies the user
// when the underlying data has changed
// It starts a goroutine which polls the API for changes
func (r *Repository) Sync() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		for {
			oldHash := r.StateHash
			r.Log.V(1).Info("start sync")
			err := r.Update()
			if err != nil {
				r.Log.Error(err, "error syncing with harbor")
			}
			if r.StateHash != oldHash {
				r.Log.V(1).Info("harbor repo state changed")
				c <- struct{}{}
			}

			r.Log.V(1).Info("end sync")
			<-time.After(r.PollInterval)
		}
	}()
	return c
}

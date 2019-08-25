package harbor

import (
	"bytes"
	"encoding/json"
	"net/url"
)

// ListProjects returns all projects
func (c *Client) ListProjects() ([]Project, error) {

	var allProjects []Project
	next := "/api/projects?page=1&page_size=10"

	for {
		var body []byte
		var err error

		nextURL, err := url.ParseRequestURI(next)
		if err != nil {
			return allProjects, err
		}
		u := c.APIBaseURL.ResolveReference(nextURL)
		body, next, err = c.pagniatedRequest(u.String())
		if err != nil {
			return allProjects, err
		}
		var projects []Project
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&projects)
		if err != nil {
			return allProjects, err
		}
		allProjects = append(allProjects, projects...)
		if next == "" {
			break
		}
	}

	return allProjects, nil
}

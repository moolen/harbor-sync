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
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/tomnomnom/linkheader"
)

// Client implements the harbor.API interface. Each func call issues a HTTP requsts to harbor
type Client struct {
	APIBaseURL *url.URL
	APIPath    string
	Username   string
	Password   string
	UserAgent  string
	HTTPClient *http.Client
	// we need to keep track of robot accounts that were
	// returned previously to get proper metrics (e.g. robot account gets deleted)
	mu                *sync.Mutex
	lastRobotAccounts map[string][]string
	debug             bool
}

// New constructs a new harbor API client
func New(baseurl, apiPath, username, password string, skipVerifyTLS, debug bool) (*Client, error) {
	if baseurl == "" {
		return nil, fmt.Errorf("API baseurl can not be empty")
	}

	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password must be set")
	}

	parsedBaseURL, err := url.Parse(baseurl)
	if err != nil {
		return nil, err
	}

	c := &http.Client{}

	if skipVerifyTLS {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c.Transport = tr
	}

	return &Client{
		APIBaseURL:        parsedBaseURL,
		APIPath:           apiPath,
		Username:          username,
		Password:          password,
		UserAgent:         "harbor-sync",
		HTTPClient:        c,
		mu:                &sync.Mutex{},
		debug:             debug,
		lastRobotAccounts: make(map[string][]string),
	}, nil
}

// BaseURL returns the base url for accessing harbor
func (c *Client) BaseURL() string {
	return c.APIBaseURL.String()
}

func (c *Client) newRequest(method string, path string, body io.Reader) (*http.Response, error) {
	// if relative path: prepend API path segment
	// CAUTION: pagination gives us an absolute path already
	if !strings.HasPrefix(path, "/") {
		path = c.APIPath + path
	}
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	relURL := c.APIBaseURL.ResolveReference(u)
	req, err := http.NewRequest(method, relURL.String(), body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	start := time.Now()
	code := 0
	defer func() {
		httpDuration := time.Since(start)
		harborAPIRequestsHistogram.WithLabelValues(fmt.Sprintf("%d", code), method, u.Path).Observe(httpDuration.Seconds())
	}()
	log.Debugf("issuing request: %s %s", req.Method, req.URL.String())
	if c.debug {
		rd, _ := httputil.DumpRequestOut(req, true)
		log.Debugf(string(rd))
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if c.debug {
		rd, _ := httputil.DumpResponse(resp, true)
		log.Debugf(string(rd))
	}
	code = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return resp, nil
}

func (c *Client) paginatedRequest(path string) ([]byte, string, error) {
	resp, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bodyBytes, "", err
	}
	links := linkheader.Parse(resp.Header.Get("Link"))
	for _, link := range links {
		if link.Rel == "next" {
			return bodyBytes, link.URL, nil
		}
	}
	return bodyBytes, "", nil
}

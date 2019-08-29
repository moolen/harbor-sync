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
	"net/url"

	"github.com/tomnomnom/linkheader"
)

// Client implements the harbor.API interface. Each func call issues a HTTP requsts to harbor
type Client struct {
	APIBaseURL *url.URL
	Username   string
	Password   string
	UserAgent  string
	HTTPClient *http.Client
}

// New constructs a new harbor API client
func New(baseurl, username, password string) (*Client, error) {
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
	// FIXME / testing
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{
		Transport: tr,
	}
	return &Client{
		APIBaseURL: parsedBaseURL,
		Username:   username,
		Password:   password,
		UserAgent:  "harbor-sync",
		HTTPClient: c,
	}, nil
}

// BaseURL returns the base url for accessing harbor
func (c *Client) BaseURL() string {
	return c.APIBaseURL.String()
}

func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}

func (c *Client) pagniatedRequest(url string) ([]byte, string, error) {
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.HTTPClient.Do(req)
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

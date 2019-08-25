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

type Client struct {
	APIBaseURL *url.URL
	Username   string
	Password   string
	UserAgent  string
	HTTPClient *http.Client

	// cache the projects and robot account responses
	// so we don't have to request them every time a SyncConfig changes
	ProjectCache      []Project
	RobotAccountCache []Robot
}

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

func (c *Client) BaseURL() string {
	return c.APIBaseURL.String()
}

func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	//
	//dumped, _ := httputil.DumpRequest(req, true)
	//fmt.Printf("%s\n", dumped)
	//
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

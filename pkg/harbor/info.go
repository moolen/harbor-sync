package harbor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

type SystemInfoResponse struct {
	WithNotary                 bool   `json:"with_notary"`
	WithClair                  bool   `json:"with_clair"`
	WithAdmiral                bool   `json:"with_admiral"`
	AdmiralEndpoint            string `json:"admiral_endpoint"`
	RegistryURL                string `json:"registry_url"`
	ExternalURL                string `json:"external_url"`
	AuthMode                   string `json:"auth_mode"`
	ProjectCreationRestriction string `json:"project_creation_restriction"`
	SelfRegistration           bool   `json:"self_registration"`
	HasCARoot                  bool   `json:"has_ca_root"`
	HarborVersion              string `json:"harbor_version"`
	NextScalAll                bool   `json:"next_scan_all"`
	ClairVulnerabilityStatus   bool   `json:"clair_vulnerability_status"`
}

type ClairVulnerabilityStatus struct {
	OverallLastUpdate int                        `json:"overall_last_update"`
	Details           []ClairVulnerabilityDetail `json:"details"`
}

type ClairVulnerabilityDetail struct {
	Namespace  string `json:"namespace"`
	LastUpdate int    `json:"last_update"`
}

func (c *Client) SystemInfo() (*SystemInfoResponse, error) {
	infoURL, err := url.ParseRequestURI("/api/systeminfo")
	if err != nil {
		return nil, err
	}
	u := c.APIBaseURL.ResolveReference(infoURL)
	req, err := c.newRequest("GET", u.String(), nil)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected HTTP response status: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var systemInfo SystemInfoResponse
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&systemInfo)
	return &systemInfo, err
}

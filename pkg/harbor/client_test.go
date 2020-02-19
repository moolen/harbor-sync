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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInfo(t *testing.T) {
	var response string
	srv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(response))
	}))
	_, err := New(srv.URL, "", "", false)
	defer srv.Close()

	if err == nil {
		t.Errorf("client should not be constructed without user/password")
	}
	c, err := New(srv.URL, "foo", "bar", false)
	if err != nil {
		t.Fail()
	}
	// info
	response = `{"harbor_version":"1.2.3"}`
	info, err := c.SystemInfo()
	if err != nil {
		t.Fail()
	}
	if info.HarborVersion != "1.2.3" {
		t.Errorf("incorrect harbor version returned")
	}
}

func TestProjects(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		if req.Form.Get("page") == "1" {
			res.Header().Set("Link", fmt.Sprintf(`<%s?page=2>; rel="next"`, srv.URL))
			res.Write([]byte(`[{"name":"project_1"},{"name":"project_2"}]`))
			return
		}
		if req.Form.Get("page") == "2" {
			res.Write([]byte(`[{"name":"project_3"},{"name":"project_4"}]`))
			return
		}
	}))
	defer srv.Close()
	c, err := New(srv.URL, "foo", "bar", false)
	if err != nil {
		t.Fail()
	}
	projects, err := c.ListProjects()
	if err != nil {
		t.Fail()
	}
	if len(projects) != 4 {
		t.Errorf("incorrect number of projects returned")
	}
}

func TestRobots(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(201)
		res.Write([]byte(`{"name":"foo","token":"bar"}`))
	}))
	defer srv.Close()
	c, err := New(srv.URL, "foo", "bar", false)
	if err != nil {
		t.Fail()
	}
	robot, err := c.CreateRobotAccount("foo", false, Project{Name: "example"})
	fmt.Printf("vals: %#v %#v", robot, err)
	if err != nil {
		t.FailNow()
	}
	if robot.Name != "foo" {
		t.Fail()
	}
	if robot.Token != "bar" {
		t.Fail()
	}
}

// fake-harbor-api implements a subset of the goharbor rest api
// its intended use-case is e2e testing with harbor-sync
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/moolen/harbor-sync/pkg/harbor"
)

type state struct {
	mu   *sync.Mutex
	data harborData
}

type harborData struct {
	Systeminfo harbor.SystemInfoResponse
	Projects   []harbor.Project
	Robots     map[string][]harbor.Robot
}

var globalState *state
var globalRobotID = 0

func main() {
	globalState = &state{
		mu: &sync.Mutex{},
		data: harborData{
			Systeminfo: harbor.SystemInfoResponse{
				HarborVersion: "1.10.1",
			},
			Projects: []harbor.Project{},
			Robots:   map[string][]harbor.Robot{},
		},
	}

	r := mux.NewRouter()
	r.HandleFunc("/_update/systeminfo", configUpdateSysteminfo).Methods("POST")
	r.HandleFunc("/_update/projects", configUpdateProjects).Methods("POST")
	r.HandleFunc("/_update/robots", configUpdateRobots).Methods("POST")
	r.HandleFunc("/api/systeminfo", systemInfoHandler).Methods("GET")
	r.HandleFunc("/api/projects", listProjectsHandler).Methods("GET")
	r.HandleFunc("/api/projects/{project_id}/robots", getRobotsHandler).Methods("GET")
	r.HandleFunc("/api/projects/{project_id}/robots", createRobotHandler).Methods("POST")
	r.HandleFunc("/api/projects/{project_id}/robots/{robot_id}", deleteRobotHandler).Methods("DELETE")
	r.Use(middleware)
	http.Handle("/", handlers.LoggingHandler(os.Stdout, r))
	log.Printf("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		globalState.mu.Lock()
		defer globalState.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		log.Printf("%q", dump)
		next.ServeHTTP(w, r)
	})
}

func configUpdateSysteminfo(w http.ResponseWriter, r *http.Request) {
	var data harbor.SystemInfoResponse
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("err decoding json: %s", err)
		return
	}
	globalState.data.Systeminfo = data
	w.WriteHeader(http.StatusOK)
}

func configUpdateProjects(w http.ResponseWriter, r *http.Request) {
	var data []harbor.Project
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("err decoding json: %s", err)
		return
	}
	globalState.data.Projects = data
	w.WriteHeader(http.StatusOK)
}

func configUpdateRobots(w http.ResponseWriter, r *http.Request) {
	var data map[string][]harbor.Robot
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("err decoding json: %s", err)
		return
	}
	globalState.data.Robots = data
	w.WriteHeader(http.StatusOK)
}

func systemInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(globalState.data.Systeminfo)
	if err != nil {
		log.Printf("err: %s", err)
	}
}

func listProjectsHandler(w http.ResponseWriter, r *http.Request) {
	pageCounter := 0
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 5
	}
	projects := []harbor.Project{}
	for idx, project := range globalState.data.Projects {
		if idx%pageSize == 0 {
			pageCounter++
		}
		if pageCounter == page {
			projects = append(projects, project)
		}
	}

	if page < pageCounter {
		w.Header().Set("Link", linkHeader(r, page+1))
	}
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(projects)
	if err != nil {
		log.Printf("err: %s", err)
	}
}

func getRobotsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if globalState.data.Robots[vars["project_id"]] == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]string{})
		return
	}
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(globalState.data.Robots[vars["project_id"]])
	if err != nil {
		log.Printf("err: %s", err)
	}
}

func createRobotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	defer r.Body.Close()
	var rr harbor.CreateRobotRequest
	err := json.NewDecoder(r.Body).Decode(&rr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("err decoding json: %s", err)
		return
	}
	if globalState.data.Robots[vars["project_id"]] == nil {
		globalState.data.Robots[vars["project_id"]] = []harbor.Robot{}
	}

	robotName := fmt.Sprintf("robot$%s", rr.Name)
	// check if it already exists
	for _, r := range globalState.data.Robots[vars["project_id"]] {
		if r.Name == robotName {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}
	globalState.data.Robots[vars["project_id"]] = append(
		globalState.data.Robots[vars["project_id"]],
		harbor.Robot{
			CreationTime: time.Now().Format(time.RFC3339Nano),
			Name:         robotName,
			ID:           globalRobotID,
		},
	)
	globalRobotID++
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(harbor.CreateRobotResponse{
		Name:  robotName,
		Token: fmt.Sprintf("mytoken$%s$", rr.Name),
	})
	if err != nil {
		log.Printf("err: %s", err)
	}
}

func deleteRobotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if globalState.data.Robots[vars["project_id"]] == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	i := -1
	for idx, ra := range globalState.data.Robots[vars["project_id"]] {
		if strconv.Itoa(ra.ID) == vars["robot_id"] {
			i = idx
		}
	}
	if i == -1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a := globalState.data.Robots[vars["project_id"]]
	copy(a[i:], a[i+1:]) // Shift a[i+1:] left one index.
	a[len(a)-1] = harbor.Robot{}
	a = a[:len(a)-1] // Truncate slice.
	globalState.data.Robots[vars["project_id"]] = a
	w.WriteHeader(http.StatusOK)
}

func linkHeader(r *http.Request, next int) string {
	q := r.URL.Query()
	q.Set("page", strconv.Itoa(next))
	r.URL.RawQuery = q.Encode()
	return fmt.Sprintf("<http://%s%s>; rel=\"next\"", r.Host, r.URL)
}

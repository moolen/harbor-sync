package harbor

// API is the interface that is implemented by the client
// and repository type
type API interface {
	ListProjects() ([]Project, error)
	GetRobotAccounts(project Project) ([]Robot, error)
	CreateRobotAccount(name string, project Project) (*CreateRobotResponse, error)
	DeleteRobotAccount(project Project, robotID int) error
	BaseURL() string
}

type Project struct {
	ID        int             `json:"project_id"`
	Name      string          `json:"name"`
	OwnerName string          `json:"owner_name"`
	Metadata  ProjectMetadata `json:"metadata"`
}

type ProjectMetadata struct {
	Public             string `json:"public"`
	EnableContentTrust string `json:"enable_content_trust"`
	PreventVul         string `json:"prevent_vul"`
	Severity           string `json:"severity"`
	AutoScan           string `json:"auto_scal"`
}

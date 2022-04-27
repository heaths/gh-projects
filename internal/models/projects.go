package models

type Project struct {
	Type   string `json:"__typename"`
	ID     string
	Number int
	Title  string
	State  string
	URL    string
}

func (p Project) IsBeta() bool {
	return p.Type == "ProjectNext"
}

type RepositoryProjects struct {
	Repository projects
}

type projects struct {
	Projects     ProjectsNode
	ProjectsNext ProjectsNode
}

type ProjectsNode struct {
	TotalCount int
	Nodes      []Project
	PageInfo   pageInfo
}

type pageInfo struct {
	HasNextPage bool
	EndCursor   string
}

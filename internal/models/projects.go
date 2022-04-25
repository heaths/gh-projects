package models

type Project struct {
	Type   string `json:"__typename"`
	ID     string
	Number int
	Title  string
}

func (p Project) IsBeta() bool {
	return p.Type == "ProjectNext"
}

type OrganizationProjects struct {
	Organization projects
}

type RepositoryProjects struct {
	Repository projects
}

type projects struct {
	Projects     projectsNode
	ProjectsNext projectsNode
}

type projectsNode struct {
	TotalCount int
	Nodes      []Project
	PageInfo   pageInfo
}

type pageInfo struct {
	HasNextPage bool
	EndCursor   string
}

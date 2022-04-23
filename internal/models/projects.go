package models

type Project struct {
	Type   string `json:"__typename"`
	ID     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
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
	Projects struct {
		TotalCount int
		Nodes      []Project
		PageInfo   pageInfo
	}
	ProjectsNext struct {
		TotalCount int
		Nodes      []Project
		PageInfo   pageInfo
	}
}

type pageInfo struct {
	HasNextPage bool
	EndCursor   string
}

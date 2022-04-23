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
	Organization projects `json:"organization"`
}

type RepositoryProjects struct {
	Repository projects `json:"repository"`
}

type projects struct {
	Projects     projectsNode `json:"projects"`
	ProjectsNext projectsNode `json:"projectsNext"`
}

type projectsNode struct {
	TotalCount int       `json:"totalCount"`
	Nodes      []Project `json:"nodes"`
	PageInfo   pageInfo  `json:"pageInfo"`
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

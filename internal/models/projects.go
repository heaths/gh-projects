package models

import "time"

type Project struct {
	Type      string `json:"__typename"`
	ID        string
	Number    int
	Title     string
	Body      string
	Creator   *actor
	CreatedAt *time.Time
	State     string
	Public    bool
	URL       string
}

func (p Project) IsBeta() bool {
	return p.Type == "ProjectNext"
}

type RepositoryProjects struct {
	Repository projects
}

type RepositoryProject struct {
	Repository project
}

type projects struct {
	Projects     ProjectsNode
	ProjectsNext ProjectsNode
}

type project struct {
	Project     Project
	ProjectNext Project
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

type actor struct {
	Login string
}

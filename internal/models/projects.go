package models

import "time"

type Project struct {
	ID          string
	Number      int
	Title       string
	Description string `json:"shortDescription"`
	Body        string `json:"description"`
	Creator     *actor
	CreatedAt   *time.Time
	Public      bool
	URL         string
}

type RepositoryProjects struct {
	Repository projects
}

type RepositoryProject struct {
	Repository ProjectNode
}

type projects struct {
	ProjectsNext ProjectsNode
}

type ProjectNode struct {
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

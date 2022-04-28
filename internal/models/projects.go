package models

import "time"

type Project struct {
	ID          string
	Number      int
	Title       string
	Description string
	Creator     *actor
	CreatedAt   *time.Time
	Public      bool
	URL         string
}

type RepositoryProjects struct {
	Repository projects
}

type RepositoryProject struct {
	Repository project
}

type projects struct {
	ProjectsNext ProjectsNode
}

type project struct {
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

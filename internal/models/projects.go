package models

import "time"

type Project struct {
	ID          string
	Number      int
	Title       string
	Description string
	Body        string
	Creator     *actor
	CreatedAt   *time.Time
	Public      bool
	URL         string
	Items       *projectItemNode
}
type projectItemNode struct {
	TotalCount int
	Nodes      []ProjectItem
	PageInfo   pageInfo
}
type ProjectItem struct {
	ID      string
	Type    string
	Content projectItemContent
}

type projectItemContent struct {
	ID        string
	Number    int
	Title     string
	CreatedAt *time.Time
	State     string
}

type RepositoryProjects struct {
	Repository projects
}

type RepositoryProject struct {
	Repository ProjectNode
}

type projects struct {
	ProjectsV2 ProjectsNode
}

type ProjectNode struct {
	ProjectV2 Project
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

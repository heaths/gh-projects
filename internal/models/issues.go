package models

type RepositoryIssueOrPullRequest struct {
	Repository issueOrPullRequestNode
}

type issueOrPullRequestNode struct {
	IssueOrPullRequest IssueOrPullRequest
}

type IssueOrPullRequest struct {
	ID string
}

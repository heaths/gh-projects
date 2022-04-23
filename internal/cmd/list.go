package cmd

import (
	"fmt"

	"github.com/cli/go-gh"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/spf13/cobra"
)

const listOrganizationProjectsQuery = `query OrganizationProjects($owner: String!, $first: Int!, $after: String) {
	organization(login: $owner) {
		projects(first: $first, after: $after) {
			nodes {
				__typename
				id
				number
				title: name
			}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
		projectsNext(first: $first, after: $after) {
			nodes {
				__typename
				id
				number
				title
			}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
	}
}`

const listRepositoryProjectsQuery = `query RepositoryProjects($owner: String!, $name: String!, $first: Int!, $after: String) {
	repository(owner: $owner, name: $name) {
		projects(first: $first, after: $after) {
			nodes {
				__typename
				id
				number
				title: name
			}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
		projectsNext(first: $first, after: $after) {
			nodes {
				__typename
				id
				number
				title
			}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
	}
}`

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return list()
		},
	}

	return cmd
}

func list() (err error) {
	client, err := gh.GQLClient(nil)
	if err != nil {
		return
	}

	organizationProjectsData := struct {
		Data models.OrganizationProjects
	}{}

	repo, err := gh.CurrentRepository()
	if err != nil {
		return
	}

	organizationProjectsVars := map[string]interface{}{
		"owner": repo.Owner(),
		"name":  repo.Name(),
		"first": 10,
	}

	err = client.Do(listOrganizationProjectsQuery, organizationProjectsVars, &organizationProjectsData)
	if err != nil {
		return
	}

	for _, project := range organizationProjectsData.Data.Organization.Projects.Nodes {
		fmt.Printf("#%d\t%s\t%s", project.Number, project.Title, project.ID)
	}

	for _, project := range organizationProjectsData.Data.Organization.ProjectsNext.Nodes {
		fmt.Printf("#%d\t%s\t%s", project.Number, project.Title, project.ID)
	}

	return nil
}

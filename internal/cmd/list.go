package cmd

import (
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/heaths/gh-projects/internal/template"
	"github.com/spf13/cobra"
)

func NewListCmd(globalOpts *GlobalOptions) *cobra.Command {
	opts := listOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.GlobalOptions = *globalOpts
			return list(&opts)
		},
	}

	IntRangeVarP(cmd, &opts.limit, "limit", "L", 30, 1, 100, "Number of projects to return")
	cmd.Flags().StringVarP(&opts.search, "search", "S", "", "Search projects.")

	return cmd
}

type listOptions struct {
	GlobalOptions

	limit  int
	search string
}

func list(opts *listOptions) (err error) {
	clientOpts := &api.ClientOptions{
		Log: opts.Log,
	}
	client, err := gh.GQLClient(clientOpts)
	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"owner": opts.Repo.Owner(),
		"name":  opts.Repo.Name(),
		"first": opts.limit,
	}

	if opts.search != "" {
		vars["search"] = opts.search
	}

	var data models.RepositoryProjects
	var projects []models.Project
	var i, totalCount int

	for {
		err = client.Do(queryRepositoryProjectsNext, vars, &data)
		if err != nil {
			return
		}

		projectsNode := data.Repository.ProjectsNext
		if projects == nil {
			totalCount = projectsNode.TotalCount
			if totalCount == 0 {
				break
			}

			projects = make([]models.Project, totalCount)
		}

		for _, project := range projectsNode.Nodes {
			projects[i] = project
			i++
		}

		if projectsNode.PageInfo.HasNextPage {
			vars["after"] = projectsNode.PageInfo.EndCursor
		} else {
			break
		}
	}

	t, err := template.New(opts.Console)
	if err != nil {
		return
	}

	return t.Projects(projects, totalCount)
}

const queryRepositoryProjectsNext = `
query RepositoryProjectsNext($owner: String!, $name: String!, $first: Int!, $after: String, $search: String) {
	repository(owner: $owner, name: $name) {
		projectsNext(first: $first, after: $after, query: $search) {
			totalCount
			nodes {
				id
				number
				title
				description: shortDescription
				public
				createdAt
			}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
	}
}
`

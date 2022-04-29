package cmd

import (
	"os"
	"text/tabwriter"

	"github.com/cli/go-gh"
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

	cmd.Flags().StringVarP(&opts.search, "search", "S", "", "Search projects.")

	return cmd
}

type listOptions struct {
	GlobalOptions

	search string
}

func list(opts *listOptions) (err error) {
	client, err := gh.GQLClient(nil)
	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"owner": opts.Repo.Owner(),
		"name":  opts.Repo.Name(),
		"first": 30,
	}

	if opts.search != "" {
		vars["search"] = opts.search
	}

	var data models.RepositoryProjects
	var projects []models.Project
	i := 0

	for {
		err = client.Do(listRepositoryProjectsNextQuery, vars, &data)
		if err != nil {
			return
		}

		projectsNode := data.Repository.ProjectsNext
		if projects == nil {
			projects = make([]models.Project, projectsNode.TotalCount)
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	return template.Projects(w, projects)
}

const listRepositoryProjectsNextQuery = `
query RepositoryProjects($owner: String!, $name: String!, $first: Int!, $after: String, $search: String) {
	repository(owner: $owner, name: $name) {
		projectsNext(first: $first, after: $after, query: $search) {
			totalCount
			nodes {
				id
				number
				title
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

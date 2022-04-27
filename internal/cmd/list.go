package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/heaths/gh-projects/internal/models"
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
	StringEnumFlag(cmd, &opts.state, "state", "s", "open", []string{"open", "closed", "all"}, "Filter by state")

	return cmd
}

type listOptions struct {
	GlobalOptions

	search string
	state  string
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

	if opts.HasState() {
		vars["states"] = []string{strings.ToUpper(opts.state)}
	}

	projects, err := listRepositoryProjects(
		client,
		listRepositoryProjectsQuery,
		vars,
		func(repo *models.RepositoryProjects) *models.ProjectsNode {
			return &repo.Repository.Projects
		},
	)
	if err != nil {
		return err
	}

	delete(vars, "states")
	projectsNext, err := listRepositoryProjects(
		client,
		listRepositoryProjectsNextQuery,
		vars,
		func(repo *models.RepositoryProjects) *models.ProjectsNode {
			return &repo.Repository.ProjectsNext
		},
	)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	for _, project := range projects {
		fmt.Fprintf(w, "#%d\t%s\t%s\n", project.Number, project.Title, project.ID)
	}
	for _, project := range projectsNext {
		fmt.Fprintf(w, "#%d\t%s\t%s\n", project.Number, project.Title, project.ID)
	}
	w.Flush()

	return nil
}

func listRepositoryProjects(client api.GQLClient, query string, vars map[string]interface{}, nodes func(*models.RepositoryProjects) *models.ProjectsNode) ([]models.Project, error) {
	var data models.RepositoryProjects
	var projects []models.Project
	i := 0

	for {
		err := client.Do(query, vars, &data)
		if err != nil {
			return nil, err
		}

		projectsNode := nodes(&data)
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

	return projects, nil
}

const listRepositoryProjectsQuery = `
query RepositoryProjects($owner: String!, $name: String!, $first: Int!, $after: String, $search: String, $states: [ProjectState!]) {
	repository(owner: $owner, name: $name) {
		projects(first: $first, after: $after, search: $search, states: $states) {
			totalCount
			nodes {
				__typename
				id
				number
				title: name
				state
				url
			}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
	}
}
`

const listRepositoryProjectsNextQuery = `
query RepositoryProjects($owner: String!, $name: String!, $first: Int!, $after: String, $search: String) {
	repository(owner: $owner, name: $name) {
		projectsNext(first: $first, after: $after, query: $search) {
			totalCount
			nodes {
				__typename
				id
				number
				title
				url
			}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
	}
}
`

func (opts listOptions) HasState() bool {
	return opts.state != "all"
}

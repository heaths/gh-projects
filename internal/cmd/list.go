package cmd

import (
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/heaths/gh-projects/internal/template"
	"github.com/spf13/cobra"
)

const (
	orderAsc  = "asc"
	orderDesc = "desc"

	sortTitle   = "title"
	sortNumber  = "number"
	sortCreated = "created"
	sortUpdated = "updated"
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

	cmd.Flags().BoolVar(&opts.all, "all", false, "List all projects for the organization or user")
	IntRangeVarP(cmd, &opts.limit, "limit", "L", 30, 1, 100, "Number of projects to return")
	StringEnumVarP(cmd, &opts.order, "order", "", orderDesc, []string{orderAsc, orderDesc}, "Order of results returned, ignored unless '--sort' flag is specified")
	cmd.Flags().StringVarP(&opts.search, "search", "S", "", "Search projects")
	StringEnumVarP(cmd, &opts.sort, "sort", "", "", []string{sortTitle, sortNumber, sortCreated, sortUpdated}, "Sort fetched results")

	return cmd
}

type listOptions struct {
	GlobalOptions

	all    bool
	limit  int
	order  string
	search string
	sort   string
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

	if opts.sort != "" {
		vars["orderBy"] = map[string]interface{}{
			"field":     projectV2Order(opts.sort),
			"direction": strings.ToUpper(opts.order),
		}
	}

	var data models.RepositoryProjects
	var projects []models.Project
	var i, totalCount int

	query := queryRepositoryProjectsV2
	if opts.all {
		query = queryRepositoryOwnerProjectsV2
	}

	for {
		err = client.Do(query, vars, &data)
		if err != nil {
			return
		}

		projectsNode := data.Repository.ProjectsV2
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

const queryRepositoryProjectsV2 = `
query RepositoryProjectsV2($owner: String!, $name: String!, $first: Int!, $after: String, $search: String, $orderBy: ProjectV2Order) {
	repository(owner: $owner, name: $name) {
		projectsV2(first: $first, after: $after, query: $search, orderBy: $orderBy) {
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

const queryRepositoryOwnerProjectsV2 = `
query RepositoryOwnerProjectsV2($owner: String!, $first: Int!, $after: String, $search: String, $orderBy: ProjectV2Order) {
	repository: repositoryOwner(login: $owner) {
		...on ProjectV2Owner {
			projectsV2(first: $first, after: $after, query: $search, orderBy: $orderBy) {
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
}
`

func projectV2Order(sort string) string {
	switch sort {
	case "created":
		return "CREATED_AT"
	case "updated":
		return "UPDATED_AT"
	default:
		return strings.ToUpper(sort)
	}
}

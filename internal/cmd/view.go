package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/heaths/gh-projects/internal/template"
	"github.com/spf13/cobra"
)

func NewViewCmd(globalOpts *GlobalOptions) *cobra.Command {
	opts := viewOptions{}
	cmd := &cobra.Command{
		Use:   "view <number>",
		Short: "View a project",
		Long: heredoc.Doc(`
			View information about a project and its fields.

			The number argument can begin with a "#" symbol.
		`),
		Args: ProjectNumberArg(&opts.number),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.GlobalOptions = *globalOpts

			return view(&opts)
		},
	}

	cmd.Flags().BoolVar(&opts.items, "items", false, "Include drafts, issues, and pull requests")
	IntRangeVarP(cmd, &opts.limit, "limit", "L", 20, 1, 100, "Number of items to include")
	StringEnumVarP(cmd, &opts.state, "state", "s", "open", []string{"open", "closed", "merged", "all"}, "State of items to include")

	return cmd
}

type viewOptions struct {
	GlobalOptions

	number int
	items  bool
	limit  int
	state  string
}

func view(opts *viewOptions) (err error) {
	clientOpts := &api.ClientOptions{
		Log: opts.Log,
	}
	client, err := gh.GQLClient(clientOpts)
	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"owner":        opts.Repo.Owner(),
		"name":         opts.Repo.Name(),
		"number":       opts.number,
		"first":        opts.limit,
		"includeItems": opts.items,
	}

	var data models.RepositoryProject
	err = client.Do(queryRepositoryProjectV2+fragmentProjectV2Items, vars, &data)
	if err != nil {
		return
	}

	project := data.Repository.ProjectV2

	if opts.items {
		items := make([]models.ProjectItem, 0, opts.limit)
		for {
			for _, item := range data.Repository.ProjectV2.Items.Nodes {
				if equalItemState(item.Content.State, opts.state) {
					items = append(items, item)
				}
			}

			if len(items) < opts.limit && project.Items.PageInfo.HasNextPage {
				vars["after"] = project.Items.PageInfo.EndCursor
				err = client.Do(queryRepositoryProjectV2MoreItems+fragmentProjectV2Items, vars, &data)
				if err != nil {
					return
				}
			} else {
				break
			}
		}

		project.Items.Nodes = items
	}

	t, err := template.New(opts.Console)
	if err != nil {
		return
	}

	return t.Project(*project)
}

const queryRepositoryProjectV2 = `
query RepositoryProjectV2($owner: String!, $name: String!, $number: Int!, $first: Int!, $after: String, $includeItems: Boolean = false) {
	repository(name: $name, owner: $owner) {
		projectV2(number: $number) {
			id
			number
			title
			description: shortDescription
			body: readme
			creator {
				login
			}
			createdAt
			public
			url
			...items @include(if: $includeItems)
		}
	}
}
`

const queryRepositoryProjectV2MoreItems = `
query RepositoryProjectV2($owner: String!, $name: String!, $number: Int!, $first: Int!, $after: String) {
	repository(name: $name, owner: $owner) {
		projectV2(number: $number) {
			...items
		}
	}
}
`

const fragmentProjectV2Items = `
fragment items on ProjectV2 {
	items(first: $first, after: $after) {
		totalCount
		nodes {
			id
			type
			content {
				... on DraftIssue {
					title
					createdAt
				}
				... on Issue {
					number
					title
					createdAt
					state
				}
				... on PullRequest {
					number
					title
					createdAt
					state
				}
			}
		}
		pageInfo {
			hasNextPage
			endCursor
		}
	}
}
`

func equalItemState(value, state string) bool {
	switch state {
	case "all":
		return true
	case "open":
		return value == "OPEN"
	case "merged":
		return value == "MERGED"
	case "closed":
		return value == "CLOSED"
	default:
		return false
	}
}

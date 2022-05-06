package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh"
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

			switch {
			case opts.limit == 0:
				return fmt.Errorf("limit must be greater than 1")
			case opts.limit > 100:
				return fmt.Errorf("limit must be less than or equal to 100")
			}

			return view(&opts)
		},
	}

	cmd.Flags().BoolVar(&opts.items, "items", false, "Include drafts, issues, and pull requests.")
	cmd.Flags().Uint32VarP(&opts.limit, "limit", "L", 20, "Number of items to include.")

	return cmd
}

type viewOptions struct {
	GlobalOptions

	number int
	items  bool
	limit  uint32
}

func view(opts *viewOptions) (err error) {
	client, err := gh.GQLClient(nil)
	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"owner":        opts.Repo.Owner(),
		"name":         opts.Repo.Name(),
		"number":       opts.number,
		"limit":        opts.limit,
		"includeItems": opts.items,
	}

	var data models.RepositoryProject
	err = client.Do(queryRepositoryProjectNext, vars, &data)
	if err != nil {
		return
	}

	t, err := template.New(opts.Console)
	if err != nil {
		return
	}

	return t.Project(data.Repository.ProjectNext)
}

const queryRepositoryProjectNext = `
query RepositoryProjectNext($owner: String!, $name: String!, $number: Int!, $limit: Int!, $includeItems: Boolean = false) {
	repository(name: $name, owner: $owner) {
		projectNext(number: $number) {
			id
			number
			title
			description: shortDescription
			body: description
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

fragment items on ProjectNext {
	items(first: $limit) {
		totalCount
		nodes {
			id
			title
			type
			content {
				... on DraftIssue {
					createdAt
				}
				... on Issue {
					number
					createdAt
					state
				}
				... on PullRequest {
					number
					createdAt
					state
				}
			}
		}
	}
}
`

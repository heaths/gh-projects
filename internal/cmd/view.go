package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
		View information about a project including its columns.

		The number argument can begin with a "#" symbol.
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.GlobalOptions = *globalOpts

			number := strings.TrimPrefix(args[0], "#")
			if number, err := strconv.ParseUint(number, 10, 32); err != nil {
				return fmt.Errorf("invalid project number")
			} else {
				opts.number = uint32(number)
			}

			return view(&opts)
		},
	}

	cmd.Flags().BoolVar(&opts.beta, "beta", false, "The number specifies a beta project")

	return cmd
}

type viewOptions struct {
	GlobalOptions

	beta   bool
	number uint32
}

func view(opts *viewOptions) (err error) {
	client, err := gh.GQLClient(nil)
	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"owner":  opts.Repo.Owner(),
		"name":   opts.Repo.Name(),
		"number": opts.number,
		"first":  30,
	}

	query := viewRepositoryProjectQuery
	if opts.beta {
		query = viewRepositoryProjectNextQuery
	}

	var data models.RepositoryProject
	err = client.Do(query, vars, &data)
	if err != nil {
		return
	}

	var project models.Project
	if opts.beta {
		project = data.Repository.ProjectNext
	} else {
		project = data.Repository.Project
	}

	err = template.Project(os.Stdout, &project)

	return
}

const viewRepositoryProjectQuery = `
query Project($owner: String!, $name: String!, $number: Int!, $first: Int!) {
	repository(name: $name, owner: $owner) {
		project(number: $number) {
			__typename
			id
			number
			title: name
			body
			creator {
				login
			}
			createdAt
			state
			url
			columns(first: $first) {
				nodes {
					id
					name
				}
			}
		}
	}
}
`

const viewRepositoryProjectNextQuery = `
query Project($owner: String!, $name: String!, $number: Int!, $first: Int!) {
	repository(name: $name, owner: $owner) {
		projectNext(number: $number) {
			__typename
			id
			number
			title
			body: description
			creator {
				login
			}
			createdAt
			public
			url
			columns: fields(first: $first) {
				nodes {
					id
					name
				}
			}
		}
	}
}
`

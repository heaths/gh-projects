package cmd

import (
	"fmt"
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
		View information about a project and its fields.

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

	return cmd
}

type viewOptions struct {
	GlobalOptions

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

	var data models.RepositoryProject
	err = client.Do(viewRepositoryProjectNextQuery, vars, &data)
	if err != nil {
		return
	}

	t, err := template.New(opts.Console)
	if err != nil {
		return
	}

	return t.Project(data.Repository.ProjectNext)
}

const viewRepositoryProjectNextQuery = `
query Project($owner: String!, $name: String!, $number: Int!) {
	repository(name: $name, owner: $owner) {
		projectNext(number: $number) {
			id
			number
			title
			description
			creator {
				login
			}
			createdAt
			public
			url
		}
	}
}
`

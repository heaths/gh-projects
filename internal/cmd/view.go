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
		Args: projectNumber(&opts.number),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.GlobalOptions = *globalOpts

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

func listItems(client api.GQLClient, number int, opts *GlobalOptions) ([]models.ProjectItem, error) {
	vars := map[string]interface{}{
		"owner":  opts.Repo.Owner(),
		"name":   opts.Repo.Name(),
		"number": number,
		"first":  30,
	}

	var data models.RepositoryProject
	var projectItems []models.ProjectItem
	var i int
	for {
		err := client.Do(queryRepositoryProjectNextItems, vars, &data)
		if err != nil {
			return nil, err
		}

		projectItemsNode := data.Repository.ProjectNext.Items
		if projectItems == nil {
			totalCount := projectItemsNode.TotalCount
			if totalCount == 0 {
				break
			}
			projectItems = make([]models.ProjectItem, totalCount)
		}

		for _, projectItem := range projectItemsNode.Nodes {
			projectItems[i] = projectItem
			i++
		}

		if projectItemsNode.PageInfo.HasNextPage {
			vars["after"] = projectItemsNode.PageInfo.EndCursor
		} else {
			break
		}
	}

	return projectItems, nil
}

const queryRepositoryProjectNext = `
query RepositoryProjectNext($owner: String!, $name: String!, $number: Int!) {
	repository(name: $name, owner: $owner) {
		projectNext(number: $number) {
			id
			number
			title
			shortDescription
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

const queryRepositoryProjectNextItems = `
query RepositoryProjectNextItems($owner: String!, $name: String!, $number: Int!, $first: Int!, $after: String) {
	repository(owner: $owner, name: $name) {
		projectNext(number: $number) {
			items(first: $first, after: $after) {
				totalCount
				nodes {
					id
					content {
						... on Issue {
							id
							number
						}
						... on PullRequest {
							id
							number
						}
					}
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

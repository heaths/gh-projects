package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/heaths/gh-projects/internal/utils"
	"github.com/spf13/cobra"
)

func NewEditCmd(globalOpts *GlobalOptions) *cobra.Command {
	var description, body string
	var public bool
	var addIssues, removeIssues []string
	opts := editOptions{}
	cmd := &cobra.Command{
		Use:   "edit <number>",
		Short: "Edit a project",
		Long: heredoc.Doc(`
			Updates project settings, and adds or removes draft issues, issues,
			and pull requests.

			The number argument can begin with a "#" symbol.

			Pass "-" to --body to read from standard input.

			Issues and pull requests to add or remove from a project are referenced
			by their issue or pull request number for the specified repository. If a
			repository is not specified, the current repository is used.

			Issue and pull request number arguments can also begin with a "#" symbol.
		`),
		Example: heredoc.Doc(`
			# make the project private
			$ gh projects edit 1 --public=false

			# set the description and read the body from stdin
			$ gh projects edit 1 --description 'Initial Release' --body - < "EOF"
			  Ship our _initial release_!
			  EOF

			# add issues to a project referenced by the current repository
			$ gh projects edit 1 --add-issue 1 --add-issue 2
		`),
		Args: projectNumber(&opts.number),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.GlobalOptions = *globalOpts

			if cmd.Flags().Changed("description") {
				opts.description = &description
			}

			if cmd.Flags().Changed("body") {
				opts.body = &body
			}

			if cmd.Flags().Changed("public") {
				opts.public = &public
			}

			opts.addIssues = make([]uint32, len(addIssues))
			for i, issue := range addIssues {
				issue, err := parseRef(issue, "invalid issue number")
				if err != nil {
					return err
				}

				opts.addIssues[i] = issue
			}

			opts.removeIssues = make([]uint32, len(removeIssues))
			for i, issue := range removeIssues {
				issue, err := parseRef(issue, "invalid issue number")
				if err != nil {
					return err
				}

				opts.removeIssues[i] = issue
			}

			return edit(&opts)
		},
	}

	// title is required so we don't need a separate variable.
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Set the new title.")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Sets the new short description.")

	// Need to pass globalOpts.Console since opts.GlobalOptions has not yet been set.
	StdinStringVarP(cmd.Flags(), globalOpts.Console.Stdin(), &body, "body", "b", "", "Set the new body.")
	cmd.Flags().BoolVar(&public, "public", false, "Set the visibility.")

	cmd.Flags().StringSliceVar(&addIssues, "add-issue", nil, "Issues or pull requests to add.")
	cmd.Flags().StringSliceVar(&removeIssues, "remove-issue", nil, "Issues or pull requests to remove.")

	return cmd
}

type editOptions struct {
	GlobalOptions

	number      uint32
	title       string
	description *string
	body        *string
	public      *bool

	addIssues    []uint32
	removeIssues []uint32
}

func edit(opts *editOptions) (err error) {
	client, err := gh.GQLClient(nil)
	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"owner":  opts.Repo.Owner(),
		"name":   opts.Repo.Name(),
		"number": opts.number,
	}

	var projectData models.RepositoryProject
	err = client.Do(queryRepositoryProjectNextID, vars, &projectData)
	if err != nil {
		return
	}

	projectId := projectData.Repository.ProjectNext.ID
	vars["id"] = projectId
	if opts.title != "" {
		vars["title"] = opts.title
	}
	if opts.description != nil {
		vars["description"] = opts.description
	}
	if opts.body != nil {
		vars["body"] = opts.body
	}
	if opts.public != nil {
		vars["public"] = opts.public
	}

	var updatedProjectData struct {
		UpdateProjectNext models.ProjectNode
	}
	err = client.Do(mutationUpdateProjectNext, vars, &updatedProjectData)
	if err != nil {
		return
	}

	projectURL := updatedProjectData.UpdateProjectNext.ProjectNext.URL

	if len(opts.addIssues) > 0 {
		count := utils.Pluralize(len(opts.addIssues), "issue")

		opts.Console.StartProgress(fmt.Sprintf("Adding %s to %s", count, projectURL))
		err = addIssues(client, projectId, opts)
		opts.Console.StopProgress()

		if err != nil {
			return
		}

		if opts.Verbose && opts.Console.IsStdoutTTY() {
			fmt.Fprintf(opts.Console.Stdout(), "Added %s", count)
		}
	}

	if len(opts.removeIssues) > 0 {
		count := utils.Pluralize(len(opts.removeIssues), "issue")

		opts.Console.StartProgress(fmt.Sprintf("Removing %s %s", count, projectURL))
		err = removeItems(client, projectId, opts)
		opts.Console.StopProgress()

		if err != nil {
			return
		}

		if opts.Verbose && opts.Console.IsStdoutTTY() {
			fmt.Fprintf(opts.Console.Stdout(), "Removed %s", count)
		}
	}

	if opts.Console.IsStdoutTTY() {
		fmt.Fprintf(opts.Console.Stdout(), "%s\n", projectURL)
	}

	return
}

func addIssues(client api.GQLClient, projectId string, opts *editOptions) (err error) {
	vars := map[string]interface{}{
		"owner": opts.Repo.Owner(),
		"name":  opts.Repo.Name(),
		"id":    projectId,
	}

	for _, issue := range opts.addIssues {
		vars["number"] = issue

		var data models.RepositoryIssueOrPullRequest
		err = client.Do(queryRepositoryIssueOrPullRequestID, vars, &data)
		if err != nil {
			return
		}

		contentId := data.Repository.IssueOrPullRequest.ID
		vars["contentId"] = contentId

		var mutationData map[string]interface{}
		err = client.Do(mutationAddProjectNextItem, vars, &mutationData)
		if err != nil {
			return
		}
	}

	return
}

func removeItems(client api.GQLClient, projectId string, opts *editOptions) (err error) {
	items, err := listItems(client, int(opts.number), &opts.GlobalOptions)
	if err != nil {
		return
	}

	itemIds := make(map[uint32]string, len(items))
	for _, item := range items {
		itemIds[item.Content.Number] = item.ID
	}

	projectItemIds := make([]string, len(opts.removeIssues))
	for i, issue := range opts.removeIssues {
		if projectItemId, ok := itemIds[issue]; !ok {
			return fmt.Errorf("project does not reference #%d", issue)
		} else {
			projectItemIds[i] = projectItemId
		}
	}

	vars := map[string]interface{}{
		"id": projectId,
	}

	for _, itemId := range projectItemIds {
		vars["itemId"] = itemId

		var mutationData map[string]interface{}
		err = client.Do(mutationDeleteProjectNextItem, vars, &mutationData)
		if err != nil {
			return
		}
	}

	return
}

const mutationAddProjectNextItem = `
mutation AddProjectNextItem($id: ID!, $contentId: ID!) {
	addProjectNextItem(input: {projectId: $id, contentId: $contentId}) {
		projectNextItem {
			id
		}
	}
}
`

const mutationDeleteProjectNextItem = `
mutation DeleteProjectNextItem($id: ID!, $itemId: ID!) {
	deleteProjectNextItem(input: {projectId: $id, itemId: $itemId}) {
		deletedItemId
	}
}
`

const queryRepositoryIssueOrPullRequestID = `
query RepositoryIssueOrPullRequestID($owner: String!, $name: String!, $number: Int!) {
	repository(owner: $owner, name: $name) {
		issueOrPullRequest(number: $number) {
			... on Issue {
				id
			}
			... on PullRequest {
				id
			}
		}
	}
}
`

const mutationUpdateProjectNext = `
mutation UpdateProjectNext($id: ID!, $title: String, $description: String, $body: String, $public: Boolean) {
	updateProjectNext(
		input: {projectId: $id, title: $title, shortDescription: $description, description: $body, public: $public}
	) {
		projectNext {
			url
		}
	}
}
`

const queryRepositoryProjectNextID = `
query RepositoryProjectNextID($owner: String!, $name: String!, $number: Int!) {
	repository(name: $name, owner: $owner) {
		projectNext(number: $number) {
			id
		}
	}
}
`

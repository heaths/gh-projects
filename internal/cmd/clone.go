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

func NewCloneCmd(globalOpts *GlobalOptions, runFunc func(*cloneOptions) error) *cobra.Command {
	var description, body string
	var public bool
	opts := cloneOptions{}
	cmd := &cobra.Command{
		Use:   "clone <number>",
		Short: "Clone a project",
		Long: heredoc.Doc(`
			Clones a project and all its fields, allowing you to optionally override some fields.

			A new title is always required, and the visibility of the project being cloned is copied by default.
			Pass --public or --public=false to override.

			The number argument can begin with a "#" symbol.

			Pass "-" to --body to read from standard input.
			`),
		Example: heredoc.Doc(`
			# clone a project using its visibility
			$ gh projects clone 1 --title "new title"

			# override the description and read the body from stdin
			$ gh projects clone 1 --description 'Subsequent update' --body - < "EOF"
			  Ship our _subsequent update_!
			  EOF
		`),
		Args: ProjectNumberArg(&opts.number),
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

			if runFunc == nil {
				runFunc = clone
			}

			return runFunc(&opts)
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Set the new title")
	//nolint:errcheck
	cmd.MarkFlagRequired("title")

	cmd.Flags().StringVarP(&description, "description", "d", "", "Sets the new short description")

	// Need to pass globalOpts.Console since opts.GlobalOptions has not yet been set.
	StdinStringVarP(cmd, globalOpts.Console.Stdin(), &body, "body", "b", "", "Set the new body")

	cmd.Flags().BoolVar(&public, "public", false, "Set the visibility; otherwise, the visibility of the project being cloned is used")
	cmd.Flags().BoolVar(&opts.drafts, "include-drafts", false, "Include draft issues")

	return cmd
}

type cloneOptions struct {
	projectOptions

	drafts bool
}

func clone(opts *cloneOptions) (err error) {
	clientOpts := &api.ClientOptions{
		AuthToken: opts.authToken,
		Host:      opts.host,
		Log:       opts.Log,
	}
	client, err := gh.GQLClient(clientOpts)
	if err != nil {
		return
	}

	vars := map[string]interface{}{
		"owner":  opts.Repo.Owner(),
		"name":   opts.Repo.Name(),
		"number": opts.number,
		"title":  opts.title,
		"drafts": opts.drafts,
	}

	var projectData models.RepositoryProject
	err = client.Do(queryRepositoryProjectV2ID, vars, &projectData)
	if err != nil {
		return
	}

	projectID := projectData.Repository.ProjectV2.ID
	projectURL := projectData.Repository.ProjectV2.URL
	public := projectData.Repository.ProjectV2.Public

	vars["ownerId"] = projectData.Viewer.ID
	vars["projectId"] = projectID

	var copyProjectV2 struct {
		CopyProjectV2 models.ProjectNode
	}

	opts.Console.StartProgress(fmt.Sprintf("Cloning %s", projectURL))
	err = client.Do(mutationCopyProjectV2, vars, &copyProjectV2)
	if err == nil {
		// Use the cloned project visibility if not specified and public (default is private).
		if opts.public == nil && public {
			opts.public = utils.Ptr(public)
		}
		projectID = copyProjectV2.CopyProjectV2.ProjectV2.ID
		err = editProject(client, projectID, false, &opts.projectOptions)
	}
	opts.Console.StopProgress()

	if err != nil {
		return
	}

	projectURL = copyProjectV2.CopyProjectV2.ProjectV2.URL
	if opts.Console.IsStdoutTTY() {
		fmt.Fprintf(opts.Console.Stdout(), "%s\n", projectURL)
	}

	return
}

const mutationCopyProjectV2 = `
mutation CopyProjectV2($ownerId: ID!, $projectId: ID!, $title: String!, $drafts: Boolean = false) {
	copyProjectV2(
		input: {ownerId: $ownerId, projectId: $projectId, title: $title, includeDraftIssues: $drafts}
	) {
		projectV2 {
			id
			url
		}
	}
}
`

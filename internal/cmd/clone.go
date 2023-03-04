package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/spf13/cobra"
)

func NewCloneCmd(globalOpts *GlobalOptions, runFunc func(*cloneOptions) error) *cobra.Command {
	opts := cloneOptions{}
	cmd := &cobra.Command{
		Use:   "clone <number>",
		Short: "Clone a project",
		Long: heredoc.Doc(`
			Clones a project and all its fields; however, new title is required.

			The number argument can begin with a "#" symbol.
			`),
		Args: ProjectNumberArg(&opts.number),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.GlobalOptions = *globalOpts

			if runFunc == nil {
				runFunc = clone
			}

			return runFunc(&opts)
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Set the new title")
	cmd.MarkFlagRequired("title")

	cmd.Flags().BoolVar(&opts.drafts, "include-drafts", false, "Include draft issues")

	return cmd
}

type cloneOptions struct {
	GlobalOptions

	number int
	title  string
	drafts bool
}

func clone(opts *cloneOptions) (err error) {
	clientOpts := &api.ClientOptions{
		Log: opts.Log,
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

	vars["ownerId"] = projectData.Viewer.ID
	vars["projectId"] = projectData.Repository.ProjectV2.ID
	projectURL := projectData.Repository.ProjectV2.URL

	var copyProjectV2 struct {
		CopyProjectV2 models.ProjectNode
	}

	opts.Console.StartProgress(fmt.Sprintf("Cloning %s", projectURL))
	err = client.Do(mutationCopyProjectV2, vars, &copyProjectV2)
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
			url
		}
	}
}
`

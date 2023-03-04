package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/text"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/heaths/gh-projects/internal/utils"
	"github.com/heaths/go-console"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func NewEditCmd(globalOpts *GlobalOptions, runFunc func(*editOptions) error) *cobra.Command {
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

			# add multiple issues to a project referenced by the current repository
			$ gh projects edit 1 --add-issue 1 --add-issue 2

			# add multiple issues to a project and set custom fields
			$ gh projects edit 1 --add-issue 1,2 -f Status=Todo -f Iteration="Iteration 1"
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

			if addIssuesCount := len(addIssues); addIssuesCount > 0 {
				opts.addIssues = make([]int, len(addIssues))
				for i, issue := range addIssues {
					issue, err := parseNumber(issue, "invalid issue number")
					if err != nil {
						return err
					}

					opts.addIssues[i] = issue
				}
			}

			if removeIssuesCount := len(removeIssues); removeIssuesCount > 0 {
				opts.removeIssues = make([]int, len(removeIssues))
				for i, issue := range removeIssues {
					issue, err := parseNumber(issue, "invalid issue number")
					if err != nil {
						return err
					}

					opts.removeIssues[i] = issue
				}
			}

			if len(opts.fields) > 0 && len(opts.addIssues) == 0 {
				return fmt.Errorf("--field requires --add-issue")
			}

			if runFunc == nil {
				runFunc = edit
			}

			return runFunc(&opts)
		},
	}

	// title is required so we don't need a separate variable.
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Set the new title")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Sets the new short description")

	// Need to pass globalOpts.Console since opts.GlobalOptions has not yet been set.
	StdinStringVarP(cmd, globalOpts.Console.Stdin(), &body, "body", "b", "", "Set the new body")
	cmd.Flags().BoolVar(&public, "public", false, "Set the visibility")

	cmd.Flags().StringSliceVar(&addIssues, "add-issue", nil, "Issues or pull requests to add")
	cmd.Flags().StringSliceVar(&removeIssues, "remove-issue", nil, "Issues or pull requests to remove")

	StringToStringVarP(cmd, &opts.fields, "field", "f", nil, "Set field values when adding issues")

	return cmd
}

type projectOptions struct {
	GlobalOptions
	number      int
	title       string
	description *string
	body        *string
	public      *bool
}

type editOptions struct {
	projectOptions

	addIssues    []int
	removeIssues []int

	fields map[string]string

	workerCount int
}

func edit(opts *editOptions) (err error) {
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
	}

	project, err := getProject(client, vars, opts)
	if err != nil {
		return
	}

	projectID := project.ID
	projectURL := project.URL

	err = editProject(client, projectID, opts.title != "", &opts.projectOptions)
	if err != nil {
		return
	}

	if len(opts.addIssues) > 0 {
		count := text.Pluralize(len(opts.addIssues), "issue")

		opts.Console.StartProgress(fmt.Sprintf("Adding %s to %s", count, projectURL))
		err = addIssues(client, projectID, opts)
		opts.Console.StopProgress()

		if err != nil {
			return
		}

		if opts.Verbose && opts.Console.IsStdoutTTY() {
			fmt.Fprintf(opts.Console.Stdout(), "Added %s\n", count)
		}
	}

	if len(opts.removeIssues) > 0 {
		count := text.Pluralize(len(opts.removeIssues), "issue")

		opts.Console.StartProgress(fmt.Sprintf("Removing %s %s", count, projectURL))
		err = removeItems(client, projectID, opts)
		opts.Console.StopProgress()

		if err != nil {
			return
		}

		if opts.Verbose && opts.Console.IsStdoutTTY() {
			fmt.Fprintf(opts.Console.Stdout(), "Removed %s\n", count)
		}
	}

	if opts.Console.IsStdoutTTY() {
		fmt.Fprintf(opts.Console.Stdout(), "%s\n", projectURL)
	}

	return
}

func getProject(client api.GQLClient, vars map[string]interface{}, opts *editOptions) (*models.Project, error) {
	var projectData models.RepositoryProject
	err := client.Do(queryRepositoryProjectV2ID, vars, &projectData)
	if err != nil && utils.AsGQLError(err, "NOT_FOUND") == nil {
		return nil, err
	}

	if projectData.Repository.ProjectV2 != nil {
		return projectData.Repository.ProjectV2, nil
	}

	// Link the project if defined by an organization or user.
	err = client.Do(queryRepositoryOwnerProjectV2ID, vars, &projectData)
	if err != nil {
		return nil, err
	}

	if projectData.Repository.ProjectV2 != nil {
		repo := fmt.Sprintf("%s/%s", vars["owner"], vars["name"])
		linkVars := map[string]interface{}{
			"projectId":    projectData.Repository.ProjectV2.ID,
			"repositoryId": projectData.Repository.Repository.ID,
		}

		// Make sure progress shows for at least a short time or it may raise doubts.
		opts.Console.StartProgress(
			fmt.Sprintf("Linking project #%d to %q", vars["number"], repo),
			console.WithMinimum(time.Second),
		)
		err = client.Do(mutationLinkProjectV2ToRepository, linkVars, nil)
		opts.Console.StopProgress()

		if err != nil {
			return nil, fmt.Errorf("failed to link project #%d to %q: %w", vars["number"], repo, err)
		}

		return projectData.Repository.ProjectV2, nil
	}

	return nil, fmt.Errorf("project #%d not found for %s %q", vars["number"], projectData.Repository.Type, vars["owner"])
}

func editProject(client api.GQLClient, projectID string, requiresUpdate bool, opts *projectOptions) (err error) {
	vars := map[string]interface{}{
		"id": projectID,
	}

	if opts.title != "" {
		vars["title"] = opts.title
		// Pass requiredUpdate: true to force an update for the title.
	}
	if opts.description != nil {
		vars["description"] = *opts.description
		requiresUpdate = true
	}
	if opts.body != nil {
		vars["body"] = *opts.body
		requiresUpdate = true
	}
	if opts.public != nil {
		vars["public"] = *opts.public
		requiresUpdate = true
	}

	if requiresUpdate {
		err = client.Do(mutationUpdateProjectV2, vars, nil)
		if err != nil {
			return
		}
	}

	return
}

func addIssues(client api.GQLClient, projectID string, opts *editOptions) (err error) {
	var fields map[string]models.Field
	if len(opts.fields) > 0 {
		fields, err = getFields(client, opts)
		if err != nil {
			return
		}
	}

	workerCount := opts.workerCount
	if workerCount < 1 {
		workerCount = DefaultWorkerCount
	}
	if issueCount := len(opts.addIssues); workerCount > issueCount {
		workerCount = issueCount
	}

	issues := make(chan int)
	wg, ctx := errgroup.WithContext(context.Background())

	for i := 0; i < workerCount; i++ {
		wg.Go(func() error {
			vars := map[string]interface{}{
				"owner": opts.Repo.Owner(),
				"name":  opts.Repo.Name(),
				"id":    projectID,
			}

			for {
				select {
				case <-ctx.Done():
					return nil
				case n, ok := <-issues:
					if !ok {
						return nil
					}

					vars["number"] = n

					var data models.RepositoryIssueOrPullRequest
					err := client.Do(queryRepositoryIssueOrPullRequestID, vars, &data)
					if err != nil {
						return err
					}

					contentID := data.Repository.IssueOrPullRequest.ID
					vars["contentId"] = contentID

					var mutationData struct {
						AddProjectV2ItemByID struct {
							Item models.ProjectItem
						}
					}

					err = client.Do(mutationAddProjectV2Item, vars, &mutationData)
					if err != nil {
						return err
					}

					if len(fields) > 0 {
						itemID := mutationData.AddProjectV2ItemByID.Item.ID
						err = updateFields(client, projectID, itemID, fields, opts)
						if err != nil {
							return err
						}
					}
				}
			}
		})
	}

	for _, issue := range opts.addIssues {
		issues <- issue
	}

	close(issues)
	err = wg.Wait()

	return
}

func getFields(client api.GQLClient, opts *editOptions) (map[string]models.Field, error) {
	vars := map[string]interface{}{
		"owner":  opts.Repo.Owner(),
		"name":   opts.Repo.Name(),
		"number": opts.number,
	}

	var data struct {
		Repository struct {
			ProjectV2 struct {
				Fields struct {
					Nodes    []models.ProjectField
					PageInfo struct {
						HasNextPage bool
						EndCursor   string
					}
				}
			}
		}
	}

	fields := make(map[string]models.Field, len(opts.fields))
	for {
		err := client.Do(queryRepositoryProjectV2Fields, vars, &data)
		if err != nil {
			return nil, err
		}

		hasNextPage := data.Repository.ProjectV2.Fields.PageInfo.HasNextPage
		for name, value := range opts.fields {
			found := false
			for _, projectField := range data.Repository.ProjectV2.Fields.Nodes {
				if strings.EqualFold(name, projectField.Name) {
					field, err := models.NewField(projectField, value)
					if err != nil {
						return nil, err
					}

					fields[name] = *field
					found = true
					delete(opts.fields, name)
					break
				}
			}

			if !found && !hasNextPage {
				return nil, fmt.Errorf("field %q not defined", name)
			}
		}

		// Only fetch more pages if some specified fields haven't been found.
		if hasNextPage && len(opts.fields) > 0 {
			vars["after"] = data.Repository.ProjectV2.Fields.PageInfo.EndCursor
		} else {
			break
		}
	}

	return fields, nil
}

func updateFields(client api.GQLClient, projectID, itemID string, fields map[string]models.Field, opts *editOptions) error {
	vars := map[string]interface{}{
		"projectId": projectID,
		"itemId":    itemID,
	}

	// Fields should be indexed in both maps based on user-specified casing.
	for name, field := range fields {
		vars["fieldId"] = field.ID
		vars["value"] = field.Value

		var data interface{}
		err := client.Do(mutationUpdateProjectV2ItemFieldValue, vars, &data)
		if err != nil {
			return fmt.Errorf("failed to update field %q: %w", name, err)
		}
	}

	return nil
}

func removeItems(client api.GQLClient, projectID string, opts *editOptions) (err error) {
	items, err := listItems(client, int(opts.number), &opts.GlobalOptions)
	if err != nil {
		return
	}

	itemIds := make(map[int]string, len(items))
	for _, item := range items {
		itemIds[item.Content.Number] = item.ID
	}

	projectItemIDs := make([]string, len(opts.removeIssues))
	for i, issue := range opts.removeIssues {
		if projectItemID, ok := itemIds[issue]; !ok {
			return fmt.Errorf("project does not reference #%d", issue)
		} else {
			projectItemIDs[i] = projectItemID
		}
	}

	vars := map[string]interface{}{
		"id": projectID,
	}

	for _, itemID := range projectItemIDs {
		vars["itemId"] = itemID

		var mutationData map[string]interface{}
		err = client.Do(mutationDeleteProjectV2Item, vars, &mutationData)
		if err != nil {
			return
		}
	}

	return
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
		err := client.Do(queryRepositoryProjectV2Items, vars, &data)
		if err != nil {
			return nil, err
		}

		projectItemsNode := data.Repository.ProjectV2.Items
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

const mutationAddProjectV2Item = `
mutation AddProjectV2ItemById($id: ID!, $contentId: ID!) {
	addProjectV2ItemById(input: {projectId: $id, contentId: $contentId}) {
		item {
			id
		}
	}
}
`

const mutationDeleteProjectV2Item = `
mutation DeleteProjectV2Item($id: ID!, $itemId: ID!) {
	deleteProjectV2Item(input: {projectId: $id, itemId: $itemId}) {
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

const queryRepositoryProjectV2Items = `
query RepositoryProjectV2Items($owner: String!, $name: String!, $number: Int!, $first: Int!, $after: String) {
	repository(owner: $owner, name: $name) {
		projectV2(number: $number) {
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

const mutationUpdateProjectV2 = `
mutation UpdateProjectV2($id: ID!, $title: String, $description: String, $body: String, $public: Boolean) {
	updateProjectV2(
		input: {projectId: $id, title: $title, shortDescription: $description, readme: $body, public: $public}
	) {
		projectV2 {
			url
		}
	}
}
`

const queryRepositoryProjectV2ID = `
query RepositoryProjectV2ID($owner: String!, $name: String!, $number: Int!) {
	viewer {
		id
	}
	repository(name: $name, owner: $owner) {
		projectV2(number: $number) {
			id
			url
			public
		}
	}
}
`

const queryRepositoryOwnerProjectV2ID = `
query RepositoryOwnerProjectV2ID($owner: String!, $name: String!, $number: Int!) {
	repository: repositoryOwner(login: $owner) {
		repository(name: $name) {
			id
		}
		type: __typename
		... on ProjectV2Owner {
			projectV2(number: $number) {
				id
				url
			}
		}
	}
}
`

const mutationLinkProjectV2ToRepository = `
mutation LinkProjectV2ToRepository($projectId: ID!, $repositoryId: ID!) {
	linkProjectV2ToRepository(
		input: {projectId: $projectId, repositoryId: $repositoryId}
	) {
		repository {
			id
		}
	}
}
`

const queryRepositoryProjectV2Fields = `
query RepositoryProjectV2Fields($owner: String!, $name: String!, $number: Int!, $after: String) {
	repository(owner: $owner, name: $name) {
		projectV2(number: $number) {
			fields(first: 30, after: $after) {
				nodes {
					...on ProjectV2Field {
						id
						name
						dataType
					}
					...on ProjectV2IterationField {
						id
						name
						dataType
						configuration {
							iterations {
								id
								name: title
							}
						}
					}
					...on ProjectV2SingleSelectField {
						id
						name
						dataType
						options {
							id
							name
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

const mutationUpdateProjectV2ItemFieldValue = `
mutation UpdateProjectV2ItemFieldValue($projectId: ID!, $itemId: ID!, $fieldId: ID!, $value: ProjectV2FieldValue!) {
	updateProjectV2ItemFieldValue(
		input: {projectId: $projectId, itemId: $itemId, fieldId: $fieldId, value: $value}
	) {
		projectV2Item {
			id
		}
	}
}
`

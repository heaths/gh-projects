package main

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/auth"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/heaths/gh-projects/internal/cmd"
	"github.com/heaths/gh-projects/internal/logger"
	"github.com/heaths/go-console"
	"github.com/spf13/cobra"
)

func main() {
	var repoFlag string
	opts := &cmd.GlobalOptions{
		Console: console.System(),
	}
	rootCmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage repository projects",
		Long: heredoc.Doc(`
		Create, edit, and close repository projects. You can also
		move issues in and out of projects.

		Both current and beta projects are supported.
		`),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if opts.Verbose {
				opts.Log = logger.New(opts.Console, "black+h")
			}

			// Try to get the host from the specified repo.
			var repo repository.Repository
			if repoFlag != "" {
				repo, err = repository.Parse(repoFlag)
				if err != nil {
					return
				}
			}

			// Validate that the user is authenticated.
			var host string
			if repo != nil {
				host = repo.Host()
			}
			if host == "" {
				host, _ = auth.DefaultHost()
			}
			token, _ := auth.TokenForHost(host)
			if token == "" {
				return fmt.Errorf("use `gh auth login -s project` to authenticate with required scopes")
			}

			// If the repo is still unassigned, try to use the current repository.
			if repo == nil {
				repo, err = gh.CurrentRepository()
				if err != nil {
					return
				}
			}

			opts.Repo = repo
			return
		},
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().StringVarP(&repoFlag, "repo", "R", "", "Select another repository to use using the [HOST/]OWNER/REPO format.")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Show verbose output.")

	rootCmd.AddCommand(cmd.NewEditCmd(opts, nil))
	rootCmd.AddCommand(cmd.NewListCmd(opts))
	rootCmd.AddCommand(cmd.NewViewCmd(opts))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

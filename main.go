package main

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/heaths/gh-projects/internal/cmd"
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
			var repo repository.Repository
			if repoFlag != "" {
				repo, err = repository.Parse(repoFlag)
				if err != nil {
					return
				}
			} else {
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

	rootCmd.AddCommand(cmd.NewEditCmd(opts))
	rootCmd.AddCommand(cmd.NewListCmd(opts))
	rootCmd.AddCommand(cmd.NewViewCmd(opts))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

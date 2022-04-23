package main

import (
	"os"

	"github.com/heaths/gh-projects/internal/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage organization or repository projects",
	}

	rootCmd.AddCommand(cmd.NewListCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

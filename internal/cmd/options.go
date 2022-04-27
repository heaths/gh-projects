package cmd

import "github.com/cli/go-gh/pkg/repository"

type GlobalOptions struct {
	Repo    repository.Repository
	Verbose bool
}

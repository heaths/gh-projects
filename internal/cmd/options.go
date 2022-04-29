package cmd

import (
	"github.com/cli/go-gh/pkg/repository"
	"github.com/heaths/go-console"
)

type GlobalOptions struct {
	Console *console.Console

	Repo    repository.Repository
	Verbose bool
}

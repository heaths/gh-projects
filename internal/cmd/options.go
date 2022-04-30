package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cli/go-gh/pkg/repository"
	"github.com/heaths/go-console"
	"github.com/spf13/cobra"
)

type GlobalOptions struct {
	Console *console.Console

	Repo    repository.Repository
	Verbose bool
}

func projectNumber(number *uint32) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) (err error) {
		if len(args) == 0 {
			return fmt.Errorf("missing required project number")
		}

		*number, err = parseRef(args[0], "invalid project number")
		return
	}
}

func parseRef(ref, errMsg string) (uint32, error) {
	num := strings.TrimPrefix(ref, "#")
	if num, err := strconv.ParseUint(num, 10, 32); err != nil {
		return 0, fmt.Errorf("%s: %s", errMsg, ref)
	} else {
		return uint32(num), nil
	}
}

package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cli/go-gh/pkg/repository"
	"github.com/heaths/go-console"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type GlobalOptions struct {
	Console *console.Console

	Repo    repository.Repository
	Verbose bool
}

func StdinStringVarP(flags *pflag.FlagSet, stdin io.Reader, p *string, name, shorthand, value, usage string) {
	flags.VarP(newStdinValue(stdin, p, value), name, shorthand, usage)
}

type stdinValue struct {
	stdin io.Reader
	value *string
}

func newStdinValue(stdin io.Reader, p *string, value string) *stdinValue {
	val := stdinValue{
		stdin: stdin,
		value: p,
	}

	*val.value = value
	return &val
}

func (v *stdinValue) String() string {
	return string(*v.value)
}

func (v *stdinValue) Set(s string) error {
	if s == "-" {
		stdin, err := io.ReadAll(v.stdin)
		if err != nil {
			return fmt.Errorf("failed to read from STDIN: %w", err)
		}

		*v.value = string(stdin)
	} else {
		*v.value = s
	}

	return nil
}

func (v *stdinValue) Type() string {
	return "string"
}

func ProjectNumberArg(number *int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) (err error) {
		if len(args) == 0 {
			return fmt.Errorf("missing required project number")
		}

		*number, err = parseNumber(args[0], "invalid project number")
		return
	}
}

func parseNumber(number, message string) (int, error) {
	num := strings.TrimPrefix(number, "#")
	if num, err := strconv.ParseUint(num, 10, 32); err != nil {
		return 0, fmt.Errorf("%s: %s", message, number)
	} else {
		return int(num), nil
	}
}

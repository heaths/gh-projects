package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cli/go-gh/pkg/repository"
	"github.com/heaths/gh-projects/internal/utils"
	"github.com/heaths/go-console"
	"github.com/spf13/cobra"
)

type GlobalOptions struct {
	Console *console.Console
	Log     io.Writer

	Repo    repository.Repository
	Verbose bool
}

func IntRangeVarP(cmd *cobra.Command, p *int, name, shorthand string, defaultValue int, min, max int, usage string) {
	*p = defaultValue
	val := &intValue{
		value: p,
		min:   min,
		max:   max,
	}

	cmd.Flags().VarP(val, name, shorthand, fmt.Sprintf("%s: {%d <= %s <= %d}", usage, min, name, max))
}

func StdinStringVarP(cmd *cobra.Command, stdin io.Reader, p *string, name, shorthand, defaultValue, usage string) {
	*p = defaultValue
	val := &stdinValue{
		stdin: stdin,
		value: p,
	}

	cmd.Flags().VarP(val, name, shorthand, usage)
}

func StringEnumVarP(cmd *cobra.Command, p *string, name, shorthand, defaultValue string, values []string, usage string) {
	*p = defaultValue
	val := &enumValue{
		value:  p,
		values: values,
	}

	cmd.Flags().VarP(val, name, shorthand, fmt.Sprintf("%s: {%s}", usage, strings.Join(values, "|")))
	_ = cmd.RegisterFlagCompletionFunc(name, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	})
}

type intValue struct {
	value    *int
	min, max int
}

func (v *intValue) String() string {
	return strconv.Itoa(*v.value)
}

func (v *intValue) Set(s string) error {
	val64, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid value: %s", s)
	}

	val := int(val64)
	if val < v.min {
		return fmt.Errorf("value is less than %d", v.min)
	}
	if val > v.max {
		return fmt.Errorf("value is more than %d", v.max)
	}

	*v.value = val
	return nil
}

func (v *intValue) Type() string {
	return "int"
}

type stdinValue struct {
	stdin io.Reader
	value *string
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

type enumValue struct {
	value  *string
	values []string
}

func (v *enumValue) String() string {
	return *v.value
}

func (v *enumValue) Set(s string) error {
	if !utils.StringSliceContains(s, v.values) {
		return fmt.Errorf("valid values are {%s}", strings.Join(v.values, "|"))
	}
	*v.value = s
	return nil
}

func (v *enumValue) Type() string {
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

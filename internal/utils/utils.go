package utils

import (
	"strings"

	"github.com/cli/go-gh/pkg/api"
)

func Ptr[T any](v T) *T {
	return &v
}

func StringSliceContains(value string, values []string) bool {
	for _, v := range values {
		if strings.EqualFold(value, v) {
			return true
		}
	}

	return false
}

func AsGQLError(err error, code string) error {
	if err, ok := err.(api.GQLError); ok {
		for _, e := range err.Errors {
			if e.Type == code {
				return err
			}
		}
	}
	return nil
}

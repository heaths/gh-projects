package utils

import (
	"strings"
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

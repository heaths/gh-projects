package utils

import (
	"fmt"
	"strings"
)

func Pluralize(num int, thing string) string {
	if num == 1 {
		return fmt.Sprintf("%d %s", num, thing)
	}

	return fmt.Sprintf("%d %ss", num, thing)
}

func StringSliceContains(value string, values []string) bool {
	for _, v := range values {
		if strings.EqualFold(value, v) {
			return true
		}
	}

	return false
}

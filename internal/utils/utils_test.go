package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSliceContains(t *testing.T) {
	assert.True(t, StringSliceContains("b", []string{"a", "b", "c"}))
	assert.False(t, StringSliceContains("z", []string{"a", "b", "c"}))
}

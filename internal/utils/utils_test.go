package utils

import (
	"errors"
	"testing"

	"github.com/cli/go-gh/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestStringSliceContains(t *testing.T) {
	assert.True(t, StringSliceContains("b", []string{"a", "b", "c"}))
	assert.False(t, StringSliceContains("z", []string{"a", "b", "c"}))
}

func TestAsGQLError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		err   error
		wantE bool
	}{
		{
			name: "nil error",
			err:  nil,
		},
		{
			name: "invalid error",
			err:  errors.New("invalid error"),
		},
		{
			name: "wrong GQLErrorItem",
			err: api.GQLError{
				Errors: []api.GQLErrorItem{
					{
						Type: "INSUFFICIENT_SCOPES",
					},
				},
			},
		},
		{
			name: "right GQLErrorItem",
			err: api.GQLError{
				Errors: []api.GQLErrorItem{
					{
						Type: "OTHER",
					},
					{
						Type: "NOT_FOUND",
					},
				},
			},
			wantE: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AsGQLError(tt.err, "NOT_FOUND")
			if tt.wantE {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

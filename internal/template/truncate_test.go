package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxWidth int
		want     string
	}{
		{
			name:     "shorter",
			s:        "short",
			maxWidth: 10,
			want:     "short",
		},
		{
			name:     "exact",
			s:        "exact",
			maxWidth: 5,
			want:     "exact",
		},
		{
			name:     "too long",
			s:        "too long",
			maxWidth: 5,
			want:     "to...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.maxWidth, tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

package template

// Copied from https://raw.githubusercontent.com/cli/cli/e2973453b5cd77df1b246a6147bbed6b47e4ce1c/pkg/text/truncate.go,
// which I helped write. Made some minimal changes where I don't need to export functions currently.

import (
	"github.com/muesli/reflow/ansi"
	trunc "github.com/muesli/reflow/truncate"
)

const (
	ellipsis            = "..."
	minWidthForEllipsis = len(ellipsis) + 2
)

// displayWidth calculates what the rendered width of a string may be.
func displayWidth(s string) int {
	return ansi.PrintableRuneWidth(s)
}

// truncate shortens a string to fit the maximum display width.
func truncate(maxWidth int, s string) string {
	w := displayWidth(s)
	if w <= maxWidth {
		return s
	}

	tail := ""
	if maxWidth >= minWidthForEllipsis {
		tail = ellipsis
	}

	r := trunc.StringWithTail(s, uint(maxWidth), tail)
	if displayWidth(r) < maxWidth {
		r += " "
	}

	return r
}

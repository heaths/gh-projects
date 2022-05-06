package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/heaths/gh-projects/internal/utils"
)

func ago(t time.Time) string {
	var approx string
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"

	case d < time.Hour:
		approx = utils.Pluralize(int(d.Minutes()), "minute")

	case d < 24*time.Hour:
		approx = utils.Pluralize(int(d.Hours()), "hour")

	case d < 30*24*time.Hour:
		approx = utils.Pluralize(int(d.Hours())/24, "day")

	case d < 365*24*time.Hour:
		approx = utils.Pluralize(int(d.Hours())/24/30, "month")

	default:
		approx = utils.Pluralize(int(d.Hours())/24/365, "year")
	}

	return fmt.Sprintf("about %s ago", approx)
}

var renderer *glamour.TermRenderer

func markdown(isTTY func() bool) func(string) (string, error) {
	return func(text string) (string, error) {
		if isTTY() {
			if renderer == nil {
				var err error
				renderer, err = glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
				)
				if err != nil {
					return "", err
				}
			}

			return renderer.Render(text)
		}

		return text, nil
	}
}

type tableState struct {
	b  *bytes.Buffer
	tw *tabwriter.Writer

	started bool
}

func (ts *tableState) init() error {
	if ts.b == nil {
		ts.b = &bytes.Buffer{}
		ts.tw = tabwriter.NewWriter(ts.b, 0, 0, 2, ' ', 0)
	}

	if ts.started {
		return fmt.Errorf("table already started")
	}

	ts.started = true
	return nil
}

func (ts *tableState) flush() (value string, err error) {
	if !ts.started {
		return "", fmt.Errorf("table not started")
	}

	err = ts.tw.Flush()
	if err != nil {
		return
	}

	value = ts.b.String()
	ts.b.Reset()

	ts.started = false
	return
}

func tablerowFunc(ts *tableState) func(...string) (string, error) {
	err := ts.init()
	if err != nil {
		panic(err)
	}
	return func(fields ...string) (string, error) {
		_, err = fmt.Fprintf(ts.tw, "%s\n", strings.Join(fields, "\t"))
		return "", err
	}
}

func tablerenderFunc(ts *tableState) func() (string, error) {
	return func() (string, error) {
		return ts.flush()
	}
}

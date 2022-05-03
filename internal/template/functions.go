package template

import (
	"fmt"
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

package template

import (
	"fmt"
	"time"
)

func ago(t time.Time) string {
	var approx string
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"

	case d < time.Hour:
		approx = pluralize(int(d.Minutes()), "minute")

	case d < 24*time.Hour:
		approx = pluralize(int(d.Hours()), "hour")

	case d < 30*24*time.Hour:
		approx = pluralize(int(d.Hours())/24, "day")

	case d < 365*24*time.Hour:
		approx = pluralize(int(d.Hours())/24/30, "month")

	default:
		approx = pluralize(int(d.Hours())/24/365, "year")
	}

	return fmt.Sprintf("about %s ago", approx)
}

func pluralize(num int, thing string) string {
	if num == 1 {
		return fmt.Sprintf("%d %s", num, thing)
	}

	return fmt.Sprintf("%d %ss", num, thing)
}

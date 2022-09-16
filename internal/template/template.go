package template

// cSpell:ignore templ
import (
	"fmt"
	"io"
	"text/tabwriter"
	tt "text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/pkg/text"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/heaths/go-console"
)

type Template struct {
	t  *tt.Template
	w  io.Writer
	ts tableState
}

func New(c console.Console) (*Template, error) {
	templ := tt.New("")
	t := &Template{
		t: templ,
		w: c.Stdout(),
	}
	cs := c.ColorScheme()
	templ.Funcs(map[string]interface{}{
		"ago":  ago,
		"bold": cs.ColorFunc("white+b"),
		"dim":  cs.ColorFunc("white+d"),
		"color": func(style, text string) string {
			return cs.ColorFunc(style)(text)
		},
		"isTTY":    c.IsStdoutTTY,
		"markdown": markdown(c.IsStdoutTTY),
		"number": func(number int) string {
			if number != 0 {
				return cs.Green(fmt.Sprintf("#%d", number))
			}
			// Return colored empty string to use consistent column width.
			return cs.Black("")
		},
		"pluralize": text.Pluralize,
		"state": func(s string) string {
			switch s {
			case "CLOSED":
				return cs.LightBlack("closed")
			case "OPEN":
				return cs.Green("open")
			case "MERGED":
				return cs.Magenta("merged")
			default:
				// Return colored empty string to use consistent column width.
				return cs.Black("")
			}
		},
		"tablerow":    tablerowFunc(&t.ts),
		"tablerender": tablerenderFunc(&t.ts),
		"truncate":    truncate,
		"type": func(it string) string {
			switch it {
			case "DRAFT_ISSUE":
				return "Draft"
			case "ISSUE":
				return "Issue"
			case "PULL_REQUEST":
				return "PullRequest"
			default:
				return it
			}
		},
		"visibility": func(public bool) string {
			if public {
				return cs.LightBlack("public")
			}

			return cs.Yellow("private")
		},
	})

	return t, nil
}

func (t *Template) Project(project models.Project) error {
	if _, err := t.t.New("project").Parse(heredoc.Doc(`
		{{bold .Title}} {{number .Number}}{{if .Description}}
		{{.Description}}{{end}}
		{{visibility .Public}} • {{.Creator.Login}} opened {{ago .CreatedAt}}
		{{if .Body}}
		{{if isTTY}}  {{end}}{{markdown .Body}}{{end}}{{with .Items}}
		Showing {{len .Nodes}} of {{pluralize .TotalCount "item"}}

		{{range .Nodes}}{{tablerow (type .Type) (number .Content.Number) (.Content.Title | truncate 80) (state .Content.State) (ago .Content.CreatedAt | dim)}}{{end}}{{tablerender}}{{end}}{{if isTTY}}
		{{printf "View this project on GitHub: %s" .URL | dim}}{{end}}
	`)); err != nil {
		return err
	}

	w := tabwriter.NewWriter(t.w, 0, 0, 2, ' ', 0)
	defer w.Flush()

	return t.t.ExecuteTemplate(t.w, "project", project)
}

func (t *Template) Projects(projects []models.Project, totalCount int) error {
	if _, err := t.t.New("projects").Parse(heredoc.Doc(`
		{{if isTTY}}
		Showing {{len .Projects}} of {{pluralize .TotalCount "project"}}

		{{end}}{{range .Projects}}{{tablerow (number .Number) (bold .Title) (visibility .Public) (ago .CreatedAt | dim) .Description}}{{end}}{{tablerender}}
	`)); err != nil {
		return err
	}

	data := struct {
		Projects   []models.Project
		TotalCount int
	}{
		Projects:   projects,
		TotalCount: totalCount,
	}

	return t.t.ExecuteTemplate(t.w, "projects", data)
}

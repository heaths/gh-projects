package template

// cSpell:ignore templ
import (
	"io"
	"text/tabwriter"
	tt "text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/heaths/go-console"
	"github.com/heaths/go-console/pkg/colorscheme"
)

type Template struct {
	t *tt.Template
	w io.Writer
}

func New(c *console.Console) (*Template, error) {
	templ := tt.New("")

	cs := colorscheme.New(
		colorscheme.WithTTY(c.IsStdoutTTY),
	)

	templ.Funcs(map[string]interface{}{
		"ago":  ago,
		"bold": cs.ColorFunc("white+b"),
		"color": func(color, text string) string {
			return cs.ColorFunc(color)(text)
		},
		"dim": cs.ColorFunc("white+d"),
	})

	if _, err := templ.New("visibility").Parse(heredoc.Doc(`
		{{if .Public}}{{color "magenta" "PUBLIC"}}{{else}}{{color "magenta" "PRIVATE"}}{{end}}`)); err != nil {
		return nil, err
	}

	return &Template{
		t: templ,
		w: c.Stdout(),
	}, nil
}

func (t *Template) Project(project models.Project) error {
	if _, err := t.t.New("project").Parse(heredoc.Doc(`
		{{bold .Title}} #{{.Number}}
		{{template "visibility" .}} • {{.Creator.Login}} opened {{ago .CreatedAt}}
		{{if .Description}}
		  {{.Description}}
		{{end}}
		{{printf "View this project on GitHub: %s" .URL | dim}}
	`)); err != nil {
		return err
	}

	return t.t.ExecuteTemplate(t.w, "project", project)
}

func (t *Template) Projects(projects []models.Project) error {
	if _, err := t.t.New("projects").Parse(heredoc.Doc(`
		{{range .}}{{printf "#%d" .Number | color "green"}}{{"\t"}}{{.Title}}{{"\t"}}{{ago .CreatedAt | dim}}{{"\t"}}{{template "visibility" .}}{{"\t"}}{{dim .ID}}{{end}}
	`)); err != nil {
		return err
	}

	w := tabwriter.NewWriter(t.w, 0, 0, 2, ' ', 0)
	defer w.Flush()

	return t.t.ExecuteTemplate(w, "projects", projects)
}

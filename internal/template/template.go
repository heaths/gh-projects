package template

// cSpell:ignore templ
import (
	"io"
	"text/tabwriter"
	tt "text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/heaths/gh-projects/internal/models"
	"github.com/heaths/gh-projects/internal/utils"
	"github.com/heaths/go-console"
)

type Template struct {
	t *tt.Template
	w io.Writer
}

func New(c *console.Console) (*Template, error) {
	templ := tt.New("")

	cs := c.ColorScheme()
	templ.Funcs(map[string]interface{}{
		"ago":  ago,
		"bold": cs.ColorFunc("white+b"),
		"dim":  cs.ColorFunc("white+d"),
		"color": func(style, text string) string {
			return cs.ColorFunc(style)(text)
		},
		"isTTY":     c.IsStdoutTTY,
		"markdown":  markdown(c.IsStdoutTTY),
		"pluralize": utils.Pluralize,
	})

	if _, err := templ.New("id").Parse(heredoc.Doc(`
		{{printf "#%d" .Number | color "green"}}`)); err != nil {
		return nil, err
	}

	if _, err := templ.New("visibility").Parse(heredoc.Doc(`
		{{if .Public}}{{color "magenta" "Public"}}{{else}}{{color "magenta" "Private"}}{{end}}`)); err != nil {
		return nil, err
	}

	return &Template{
		t: templ,
		w: c.Stdout(),
	}, nil
}

func (t *Template) Project(project models.Project) error {
	if _, err := t.t.New("project").Parse(heredoc.Doc(`
		{{bold .Title}} {{template "id" .}}{{if .Description}}
		{{.Description}}{{end}}
		{{template "visibility" .}} • {{.Creator.Login}} opened {{ago .CreatedAt}}
		{{if .Body}}
		{{if isTTY}}  {{end}}{{markdown .Body}}{{end}}{{if isTTY}}
		{{printf "View this project on GitHub: %s" .URL | dim}}{{end}}
	`)); err != nil {
		return err
	}

	return t.t.ExecuteTemplate(t.w, "project", project)
}

func (t *Template) Projects(projects []models.Project, totalCount int) error {
	if _, err := t.t.New("projects").Parse(heredoc.Doc(`
		{{if isTTY}}
		Showing {{len .Projects}} of {{pluralize .TotalCount "project"}}

		{{end}}{{range .Projects}}{{template "id" .}}{{"\t"}}{{bold .Title}}{{"\t"}}{{template "visibility" .}}{{"\t"}}{{ago .CreatedAt | dim}}{{"\t"}}{{.Description}}{{end}}
	`)); err != nil {
		return err
	}

	w := tabwriter.NewWriter(t.w, 0, 0, 2, ' ', 0)
	defer w.Flush()

	data := struct {
		Projects   []models.Project
		TotalCount int
	}{
		Projects:   projects,
		TotalCount: totalCount,
	}

	return t.t.ExecuteTemplate(w, "projects", data)
}

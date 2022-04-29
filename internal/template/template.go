package template

// cSpell:ignore templ
import (
	"io"
	tt "text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/heaths/gh-projects/internal/models"
)

type Template struct {
	t *tt.Template
	w io.Writer
}

func New(w io.Writer) (*Template, error) {
	templ, err := tt.New("visibility").Parse(`{{if .Public}}PUBLIC{{else}}PRIVATE{{end}}`)
	if err != nil {
		return nil, err
	}

	templ.Funcs(map[string]interface{}{
		"ago": ago,
	})

	return &Template{t: templ, w: w}, nil
}

func (t *Template) Project(project models.Project) error {
	t.t.New("project").Parse(heredoc.Doc(`
	{{.Title}} #{{.Number}}
	{{template "visibility" .}} • {{.Creator.Login}} opened {{ago .CreatedAt}}
	{{if .Description}}
	  {{.Description}}
	{{end}}
	View this project on GitHub: {{.URL}}
	`))

	return t.t.ExecuteTemplate(t.w, "project", project)
}

func (t *Template) Projects(projects []models.Project) error {
	t.t.New("projects").Parse(heredoc.Doc(`
	{{range .}}#{{.Number}}{{"\t"}}{{.Title}}{{"\t"}}{{ago .CreatedAt}}{{"\t"}}{{template "visibility" .}}{{"\t"}}{{.ID}}{{end}}
	`))

	return t.t.ExecuteTemplate(t.w, "projects", projects)
}

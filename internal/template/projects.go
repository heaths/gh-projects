package template

import (
	"io"
	"text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/heaths/gh-projects/internal/models"
)

var templ *template.Template

func init() {
	templ = template.Must(template.New("public").Parse(heredoc.Doc(`
	{{if .Public}}PUBLIC{{else}}PRIVATE{{end}}`)))

	template.Must(templ.New("project").Parse(heredoc.Doc(`
	{{.Title}} #{{.Number}}
	{{if .State}}{{.State}}{{else}}{{template "public" .}}{{end}} • {{.Creator.Login}} opened {{.CreatedAt}}
	{{if .Body}}

	{{.Body}}

	{{end}}
	View this project on GitHub: {{.URL}}
	`)))
}

func Project(w io.Writer, project *models.Project) error {
	return templ.ExecuteTemplate(w, "project", project)
}

package template

// cSpell:ignore templ

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
	{{template "public" .}} • {{.Creator.Login}} opened {{.CreatedAt}}
	{{if .Description}}
	{{.Description}}
	{{end}}
	View this project on GitHub: {{.URL}}
	`)))

	template.Must(templ.New("projects").Parse(heredoc.Doc(`
	{{range .}}#{{.Number}}{{"\t"}}{{.Title}}{{"\t"}}{{.CreatedAt}}{{"\t"}}{{template "public" .}}{{"\t"}}{{.ID}}{{end}}
	`)))
}

func Project(w io.Writer, project *models.Project) error {
	return templ.ExecuteTemplate(w, "project", project)
}

func Projects(w io.Writer, projects []models.Project) error {
	return templ.ExecuteTemplate(w, "projects", projects)
}

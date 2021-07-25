package core

import (
	"html/template"
	"io"
)

// Report structure
type Report struct {
	Session  *Session
	Template string
}

// Render the report template
func (r *Report) Render(dest io.Writer) error {
	funcMap := template.FuncMap{
		"json": func(json string) template.JS {
			return template.JS(json)
		},
	}
	tmpl, err := template.New("Aquasily Report").Funcs(funcMap).Parse(r.Template)
	if err != nil {
		return err
	}
	err = tmpl.Execute(dest, r.Session)
	if err != nil {
		return err
	}
	return nil
}

// NewReport returns a new report
func NewReport(s *Session, templ string) *Report {
	return &Report{
		Session:  s,
		Template: templ,
	}
}

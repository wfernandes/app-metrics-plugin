package views

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/wfernandes/app-metrics-plugin/pkg/agent"
)

type ViewOpt func(*View)

func WithWriter(w io.Writer) ViewOpt {
	return func(v *View) {
		v.writer = w
	}
}

func WithTemplate(b *template.Template) ViewOpt {
	return func(v *View) {
		v.tmpl = b
	}
}

type View struct {
	writer io.Writer
	tmpl   *template.Template
}

func New(opts ...ViewOpt) *View {
	v := &View{
		writer: os.Stdout,
		tmpl:   buildDefaultTemplate(),
	}

	for _, o := range opts {
		o(v)
	}

	return v
}

func (v *View) Present(m []agent.InstanceMetric) error {

	err := v.tmpl.Execute(v.writer, m)
	if err != nil {
		return fmt.Errorf("unable to render template %s: %s", v.tmpl.Name(), err)
	}
	return nil
}

func buildDefaultTemplate() *template.Template {
	t := template.New("default")
	// TODO: Ignoring this error for now
	t, _ = t.Parse(`
{{- range .}}
Instance: {{.Instance}}
{{ if .Metrics -}}
Metrics:
  {{- range $k, $v := .Metrics}}
  {{print $k}}: {{print $v -}}
  {{end -}}
{{else -}}
Error: {{.Error}}
{{end }}
{{end}}`)

	return t
}

package views

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/wfernandes/apps-metrics-plugin/pkg/agent"
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

func (v *View) Present(m []agent.MetricOuput) error {

	err := v.tmpl.Execute(v.writer, m)
	if err != nil {
		return fmt.Errorf("unable to render template %s: %s", v.tmpl.Name(), err)
	}
	return nil
}

func buildDefaultTemplate() *template.Template {
	t := template.New("default")
	t = t.Funcs(template.FuncMap{"metricsParse": ParseMetrics})
	// TODO: Ignoring this error for now
	t, _ = t.Parse(`
{{- range .}}
Instance: {{.Instance}}
{{ if .Output -}}
Metrics:
  {{- range $k, $v := metricsParse .Output}}
  {{print $k}}: {{printf "%s" $v -}}
  {{end -}}
{{else -}}
Error: {{.Error}}
{{end }}
{{end}}`)

	return t
}

func ParseMetrics(s string) (map[string]json.RawMessage, error) {
	metrics := make(map[string]json.RawMessage)
	err := json.Unmarshal([]byte(s), &metrics)
	return metrics, err
}

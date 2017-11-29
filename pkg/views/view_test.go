package views_test

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/wfernandes/app-metrics-plugin/pkg/agent"
	"github.com/wfernandes/app-metrics-plugin/pkg/views"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Views", func() {
	Context("with default template", func() {
		It("displays the metrics data", func() {
			metrics := []agent.InstanceMetric{}
			err := json.Unmarshal([]byte(getMetricsOutput), &metrics)
			Expect(err).ToNot(HaveOccurred())
			buf := &bytes.Buffer{}

			v := views.New(views.WithWriter(buf))
			err = v.Present(metrics)
			Expect(err).ToNot(HaveOccurred())

			bufStr := buf.String()

			Expect(bufStr).To(ContainSubstring("Instance: 0"))
			Expect(bufStr).To(ContainSubstring("Metrics:"))
			Expect(bufStr).To(ContainSubstring("  metric.float: 123.345"))
			Expect(bufStr).To(ContainSubstring("  metric.int: 10"))
			Expect(bufStr).To(ContainSubstring("  metric.string: expvarApp"))
			Expect(bufStr).To(ContainSubstring("  metric.map: ")) // maps are unordered so `map[metric1:10 metric2:11]` prints in non-deterministic order
			Expect(bufStr).To(ContainSubstring("Instance: 1"))
			Expect(bufStr).To(ContainSubstring("Error: unable to parse response: invalid character 'p' after top-level value"))
		})

	})

	Context("with custom template", func() {
		It("display the metrics data", func() {
			metrics := []agent.InstanceMetric{}
			err := json.Unmarshal([]byte(getMetricsOutput), &metrics)
			Expect(err).ToNot(HaveOccurred())

			buf := &bytes.Buffer{}
			tmpl := template.New("test")
			tmpl, err = tmpl.Parse(`
				{{range .}}
					Instance: {{.Instance}}
					Output: {{.Metrics}}
					Error: {{.Error}}
				{{end}}
			`)
			Expect(err).ToNot(HaveOccurred())

			v := views.New(views.WithWriter(buf), views.WithTemplate(tmpl))
			err = v.Present(metrics)
			Expect(err).ToNot(HaveOccurred())

			bufStr := buf.String()
			Expect(bufStr).To(ContainSubstring("Instance: 0"))
			Expect(bufStr).To(ContainSubstring("Output: "))
			Expect(bufStr).To(ContainSubstring("Instance: 1"))
			Expect(bufStr).To(ContainSubstring("Error: unable to parse response: invalid character 'p' after top-level value"))
		})

		It("returns an error when template fails to execute", func() {
			metrics := []agent.InstanceMetric{}
			err := json.Unmarshal([]byte(getMetricsOutput), &metrics)
			Expect(err).ToNot(HaveOccurred())

			buf := &bytes.Buffer{}
			tmpl := template.New("test")
			tmpl, err = tmpl.Parse(`
				{{range .}}
					Instance: {{.Instance}}
					Output: {{.output}}
					Error: {{.Error}}
				{{end}}
			`)
			Expect(err).ToNot(HaveOccurred())

			v := views.New(views.WithWriter(buf), views.WithTemplate(tmpl))
			err = v.Present(metrics)
			Expect(err).To(HaveOccurred())
		})
	})
})

var getMetricsOutput = `
[
  {
    "Instance": 0,
    "Error": "",
    "Metrics": {
      "metric.float": 123.345,
      "metric.int": 10,
      "metric.map": {
        "metric1": 10,
        "metric2": 11
      },
      "metric.string": "expvarApp"
    }
  },
  {
    "Instance": 1,
    "Error": "unable to parse response: invalid character 'p' after top-level value",
    "Metrics": null
  }
]`

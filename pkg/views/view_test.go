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
			metrics := []agent.MetricOuput{}
			err := json.Unmarshal([]byte(jsonMetricsOutput), &metrics)
			Expect(err).ToNot(HaveOccurred())
			buf := &bytes.Buffer{}

			v := views.New(views.WithWriter(buf))
			err = v.Present(metrics)
			Expect(err).ToNot(HaveOccurred())

			bufStr := buf.String()

			Expect(bufStr).To(ContainSubstring("Instance: 0"))
			Expect(bufStr).To(ContainSubstring("Metrics:"))
			Expect(bufStr).To(ContainSubstring("  ingress.matched: 11"))
			Expect(bufStr).To(ContainSubstring("  ingress.received: 5850203"))
			Expect(bufStr).To(ContainSubstring("  notifier.dropped: 0"))
			Expect(bufStr).To(ContainSubstring("  notifier.emails.failed: 0"))
			Expect(bufStr).To(ContainSubstring("  notifier.emails.sent: 11"))
			Expect(bufStr).To(ContainSubstring("Instance: 1"))
			Expect(bufStr).To(ContainSubstring("Error: some error in getting metrics"))
		})

		It("returns an error when template fails to execute", func() {
			metrics := []agent.MetricOuput{}
			err := json.Unmarshal([]byte(badJSONMetricsOutput), &metrics)
			Expect(err).ToNot(HaveOccurred())
			buf := &bytes.Buffer{}

			v := views.New(views.WithWriter(buf))
			err = v.Present(metrics)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("with custom template", func() {
		It("display the metrics data", func() {
			metrics := []agent.MetricOuput{}
			err := json.Unmarshal([]byte(jsonMetricsOutput), &metrics)
			Expect(err).ToNot(HaveOccurred())

			buf := &bytes.Buffer{}
			tmpl := template.New("test")
			tmpl, err = tmpl.Parse(`
				{{range .}}
					Instance: {{.Instance}}
					Output: {{.Output}}
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
			Expect(bufStr).To(ContainSubstring("Error: some error in getting metrics"))
		})

		It("returns an error when template fails to execute", func() {
			metrics := []agent.MetricOuput{}
			err := json.Unmarshal([]byte(jsonMetricsOutput), &metrics)
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

var jsonMetricsOutput = `[{
"Instance": 0,
"Output": "{\"ingress.matched\":11,\"ingress.received\":5850203,\"notifier.dropped\":0,\"notifier.emails.failed\":0,\"notifier.emails.sent\":11}",
"Error": ""
},{
"Instance": 1,
"Output": "",
"Error": "some error in getting metrics"
}] `

var badJSONMetricsOutput = `[{
"Instance": 0,
"Output": "{\"ingress.matched\":,\"notifier.emails.sent\":11,}",
"Error": ""
}] `

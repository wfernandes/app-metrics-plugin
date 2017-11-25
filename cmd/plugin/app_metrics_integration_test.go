package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"

	. "code.cloudfoundry.org/cli/util/testhelpers/io"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/wfernandes/app-metrics-plugin/cmd/plugin"
)

var _ = Describe("AppsMetrics Integration", func() {

	It("returns fallback output style when presentation fails", func() {
		endpoint := "/myspecific/metrics/endpoint"
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		mux := http.NewServeMux()
		ts := httptest.NewServer(mux)
		defer ts.Close()
		mux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "this is my metrics output")
		})

		// trimming the scheme because we'll build the url back from app model
		model := buildAppModel(strings.TrimPrefix(ts.URL, "http://"), 1)
		fakeCliConnection.GetAppReturns(model, nil)

		appsMetricsPlugin := &AppsMetricsPlugin{}
		output := CaptureOutput(func() {
			appsMetricsPlugin.Run(fakeCliConnection, []string{"app-metrics", "some-app", "-endpoint", endpoint})
		})

		Expect(output).To(ContainElement("[{\"Instance\":0,\"Output\":\"this is my metrics output\",\"Error\":\"\"}]"))
	})

	It("returns default template output style", func() {
		endpoint := "/myspecific/metrics/endpoint"
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		mux := http.NewServeMux()
		ts := httptest.NewServer(mux)
		defer ts.Close()
		mux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"ingress.received": 12345,"ingress.sent": 12345}`)
		})

		// trimming the scheme because we'll build the url back from app model
		model := buildAppModel(strings.TrimPrefix(ts.URL, "http://"), 1)
		fakeCliConnection.GetAppReturns(model, nil)

		appsMetricsPlugin := &AppsMetricsPlugin{}
		output := CaptureOutput(func() {
			appsMetricsPlugin.Run(fakeCliConnection, []string{"app-metrics", "some-app", "-endpoint", endpoint})
		})

		Expect(output).To(ContainElement("Instance: 0"))
		Expect(output).To(ContainElement("Metrics:"))
		Expect(output).To(ContainElement("  ingress.received: 12345"))
		Expect(output).To(ContainElement("  ingress.sent: 12345"))
	})

	It("returns custom template output style", func() {
		// setup test server/app
		mux := http.NewServeMux()
		ts := httptest.NewServer(mux)
		defer ts.Close()
		mux.HandleFunc("/debug/metrics", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"ingress.received": 222}`)
		})

		// setup fake app model with multiple app instances
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		model := buildAppModel(strings.TrimPrefix(ts.URL, "http://"), 2)
		fakeCliConnection.GetAppReturns(model, nil)

		// setup temporary template file
		content := []byte(buildTemplate())
		tmpfile, err := ioutil.TempFile("", "example")
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(tmpfile.Name())
		_, err = tmpfile.Write(content)
		Expect(err).ToNot(HaveOccurred())
		err = tmpfile.Close()
		Expect(err).ToNot(HaveOccurred())

		plugin := &AppsMetricsPlugin{}
		output := CaptureOutput(func() {
			plugin.Run(fakeCliConnection, []string{"app-metrics", "some-app", "-template", tmpfile.Name()})
		})
		Expect(output).To(HaveLen(3))
		Expect(output).To(ContainElement(`0 {"ingress.received":222}`))
		Expect(output).To(ContainElement(`1 {"ingress.received":222}`))
	})

	It("prints error if unable to parse template files", func() {
		// setup test server/app
		mux := http.NewServeMux()
		ts := httptest.NewServer(mux)
		defer ts.Close()
		mux.HandleFunc("/debug/metrics", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"ingress.received": 222}`)
		})

		// setup fake app model with multiple app instances
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		model := buildAppModel(strings.TrimPrefix(ts.URL, "http://"), 2)
		fakeCliConnection.GetAppReturns(model, nil)

		plugin := &AppsMetricsPlugin{}
		output := CaptureOutput(func() {
			plugin.Run(fakeCliConnection, []string{"app-metrics", "some-app", "-template", "/some/file/path"})
		})
		Expect(output).To(HaveLen(3))
		Expect(output[0]).To(Equal("FAILED"))
		Expect(output[1]).To(ContainSubstring("unable to parse template files"))
	})
})

func buildTemplate() string {
	return `
{{- range .}}
{{- .Instance}} {{.Output -}} {{.Error}}
{{ end -}}`
}

func buildAppModel(host string, runningInstances int) plugin_models.GetAppModel {
	m := plugin_models.GetAppModel{
		Guid:             "some-app-guid",
		RunningInstances: runningInstances,
		Instances:        []plugin_models.GetApp_AppInstanceFields{},
		Routes: []plugin_models.GetApp_RouteSummary{
			{
				Domain: plugin_models.GetApp_DomainFields{
					Name: host,
				},
			},
		},
	}

	for i := 0; i < runningInstances; i++ {
		m.Instances = append(m.Instances, plugin_models.GetApp_AppInstanceFields{State: "running"})
	}
	return m
}

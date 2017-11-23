package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"

	. "code.cloudfoundry.org/cli/util/testhelpers/io"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/wfernandes/apps-metrics-plugin/cmd/plugin"
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
			appsMetricsPlugin.Run(fakeCliConnection, []string{"apps-metrics", "some-app", "-endpoint", endpoint})
		})

		Expect(output).To(ContainElement("[{\"Instance\":0,\"Output\":\"this is my metrics output\",\"Error\":\"\"}]"))
	})

	It("returns templated output style", func() {
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
			appsMetricsPlugin.Run(fakeCliConnection, []string{"apps-metrics", "some-app", "-endpoint", endpoint})
		})

		Expect(output).To(ContainElement("Instance: 0"))
		Expect(output).To(ContainElement("Metrics:"))
		Expect(output).To(ContainElement("  ingress.received: 12345"))
		Expect(output).To(ContainElement("  ingress.sent: 12345"))
	})
})

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

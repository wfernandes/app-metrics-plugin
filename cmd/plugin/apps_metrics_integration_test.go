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
		model := buildAppModel(strings.TrimPrefix(ts.URL, "http://"))
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
		model := buildAppModel(strings.TrimPrefix(ts.URL, "http://"))
		fakeCliConnection.GetAppReturns(model, nil)

		appsMetricsPlugin := &AppsMetricsPlugin{}
		output := CaptureOutput(func() {
			appsMetricsPlugin.Run(fakeCliConnection, []string{"apps-metrics", "some-app", "-endpoint", endpoint})
		})

		Expect(output).To(ContainElement("\tInstance: 0"))
		Expect(output).To(ContainElement("\tMetrics:"))
		Expect(output).To(ContainElement("\t  ingress.received: 12345"))
		Expect(output).To(ContainElement("\t  ingress.sent: 12345"))
	})
})

func buildAppModel(host string) plugin_models.GetAppModel {
	return plugin_models.GetAppModel{
		Guid: "some-app-guid",
		Instances: []plugin_models.GetApp_AppInstanceFields{
			{
				State: "running",
			},
		},
		Routes: []plugin_models.GetApp_RouteSummary{
			{
				Domain: plugin_models.GetApp_DomainFields{
					Name: host,
				},
			},
		},
	}
}

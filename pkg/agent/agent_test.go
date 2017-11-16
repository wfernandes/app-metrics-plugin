package agent_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/wfernandes/apps-metrics-plugin/pkg/agent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agent", func() {

	var (
		mux *http.ServeMux
		ts  *httptest.Server
	)

	It("makes a GET request to the app metrics endpoint", func() {
		mux = http.NewServeMux()
		ts = httptest.NewServer(mux)
		defer ts.Close()
		request := &http.Request{}
		mux.HandleFunc("/debug/metrics", func(w http.ResponseWriter, r *http.Request) {
			request = r
		})
		fakeApp := &plugin_models.GetAppModel{
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: ts.URL,
					},
				},
			},
		}
		a := agent.New(fakeApp, "/debug/metrics")

		_, err := a.GetMetrics()

		Expect(err).ToNot(HaveOccurred())
		Expect(request.Method).To(Equal(http.MethodGet))
		tsUrl, _ := url.Parse(ts.URL)
		Expect(request.Host).To(Equal(tsUrl.Host))
	})

	It("returns error upon unsuccessful request", func() {
		fakeApp := &plugin_models.GetAppModel{
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "",
					},
				},
			},
		}
		a := agent.New(fakeApp, "/whatever")
		_, err := a.GetMetrics()
		Expect(err).To(HaveOccurred())
	})

	It("returns response body upon successful request", func() {
		mux = http.NewServeMux()
		ts = httptest.NewServer(mux)
		defer ts.Close()
		request := &http.Request{}
		mux.HandleFunc("/debug/metrics", func(w http.ResponseWriter, r *http.Request) {
			request = r
			fmt.Fprint(w, "something")
		})
		fakeApp := &plugin_models.GetAppModel{
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: ts.URL,
					},
				},
			},
		}
		a := agent.New(fakeApp, "/debug/metrics")

		output, err := a.GetMetrics()

		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(Equal("something"))
	})

})

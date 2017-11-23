package agent_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/wfernandes/apps-metrics-plugin/pkg/agent"
)

var _ = Describe("Agent", func() {
	It("cancels requests for slow endpoints and returns error in metric output", func() {
		mux := http.NewServeMux()
		ts := httptest.NewServer(mux)
		defer ts.Close()
		mux.HandleFunc("/debug/metrics", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-CF-APP-INSTANCE") == "some-app-guid:1" {
				select {
				case <-time.After(10 * time.Second):
				case <-r.Context().Done():
				}
			}
			fmt.Fprintf(w, `{"ingress.received": 12345,"ingress.sent": 12345}`)
		})
		model := buildAppModel(strings.TrimPrefix(ts.URL, "http://"), 3)
		// use a client with very low tolerance for slow requests
		httpClient := &http.Client{Timeout: 100 * time.Millisecond}

		a := agent.New(&model, agent.WithClient(httpClient))
		output, err := a.GetMetrics()
		Expect(err).ToNot(HaveOccurred())
		Eventually(output).Should(HaveLen(3))
		for _, m := range output {
			if m.Instance == 1 {
				Expect(m.Output).To(BeEmpty())
				Expect(m.Error).ToNot(BeEmpty())
				continue
			}
			Expect(m.Output).ToNot(BeEmpty())
			Expect(m.Error).To(BeEmpty())
		}
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

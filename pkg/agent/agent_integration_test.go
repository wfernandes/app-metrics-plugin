package agent_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/wfernandes/app-metrics-plugin/pkg/agent"
	"github.com/wfernandes/app-metrics-plugin/pkg/parser"
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
		parser := parser.NewExpvar()

		a := agent.New(&model, parser, agent.WithClient(httpClient))
		output, err := a.GetMetrics(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Eventually(output).Should(HaveLen(3))
		for _, m := range output {
			if m.Instance == 1 {
				Expect(m.Metrics).To(BeEmpty())
				Expect(m.Error).ToNot(BeEmpty())
				continue
			}
			Expect(m.Metrics).ToNot(BeEmpty())
			Expect(m.Error).To(BeEmpty())
		}
	})

	It("cancels processing requests when context is cancelled and returns whatever requests it has", func() {
		mux := http.NewServeMux()
		ts := httptest.NewServer(mux)
		defer ts.Close()
		mux.HandleFunc("/debug/metrics", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-CF-APP-INSTANCE") == "some-app-guid:1" {
				select {
				case <-time.After(5 * time.Second):
				case <-r.Context().Done():
					return
				}
			}
			fmt.Fprintf(w, `{"ingress.received": 12345,"ingress.sent": 12345}`)
		})
		// trimming the scheme because we'll build the url back from app model
		model := buildAppModel(strings.TrimPrefix(ts.URL, "http://"), 2)
		httpClient := &http.Client{Timeout: 10 * time.Second}
		ctx, cancel := context.WithCancel(context.Background())
		parser := parser.NewExpvar()

		a := agent.New(&model, parser, agent.WithClient(httpClient))

		outputs := make(chan []agent.InstanceMetric, 2)
		errs := make(chan error)
		go func() {
			o, err := a.GetMetrics(ctx)
			outputs <- o
			errs <- err
		}()
		// give some time for the first request to process
		time.Sleep(time.Second)
		cancel()

		var results []agent.InstanceMetric
		Eventually(errs).Should(Receive())
		Eventually(outputs).Should(Receive(&results))
		Expect(results).To(HaveLen(1))
		Expect(results[0].Instance).To(Equal(0))
	})

	It("sorts the output by instance number", func() {
		mux := http.NewServeMux()
		ts := httptest.NewServer(mux)
		defer ts.Close()
		mux.HandleFunc("/debug/metrics", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"instance.name": %s}`, r.Header.Get("X-CF-APP-INSTANCE"))
		})
		// trimming the scheme because we'll build the url back from app model
		model := buildAppModel(strings.TrimPrefix(ts.URL, "http://"), 3)
		parser := parser.NewExpvar()

		a := agent.New(&model, parser)

		output, err := a.GetMetrics(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(HaveLen(3))
		Expect(output[0].Instance).To(Equal(0))
		Expect(output[1].Instance).To(Equal(1))
		Expect(output[2].Instance).To(Equal(2))
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

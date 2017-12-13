package parser_test

import (
	"encoding/json"

	"github.com/wfernandes/app-metrics-plugin/pkg/parser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prometheus", func() {

	It("parses metrics into a map", func() {
		p := parser.NewPrometheus()
		metrics, err := p.Parse([]byte(prometheusOutput))
		Expect(err).ToNot(HaveOccurred())
		Expect(metrics).ToNot(BeEmpty())
		b, err := json.Marshal(metrics)
		Expect(err).ToNot(HaveOccurred())
		Expect(b).To(MatchJSON(promJSON))
	})

	It("returns error and empty map if error occurs in parsing", func() {
		p := parser.NewPrometheus()
		metrics, err := p.Parse([]byte("bla"))
		Expect(err).To(HaveOccurred())
		Expect(metrics).To(BeEmpty())
	})

})

var prometheusOutput = `# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
go_gc_duration_seconds{quantile="0.75"} 0
go_gc_duration_seconds{quantile="1"} 0
go_gc_duration_seconds_sum 0
go_gc_duration_seconds_count 0
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 6
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.9.1"} 1
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 483504
`

var promJSON = `{"go_gc_duration_seconds":{"name":"go_gc_duration_seconds","help":"A summary of the GC invocation durations.","type":2,"metric":[{"summary":{"sample_count":0,"sample_sum":0,"quantile":[{"quantile":0,"value":0},{"quantile":0.25,"value":0},{"quantile":0.5,"value":0},{"quantile":0.75,"value":0},{"quantile":1,"value":0}]}}]},"go_goroutines":{"name":"go_goroutines","help":"Number of goroutines that currently exist.","type":1,"metric":[{"gauge":{"value":6}}]},"go_info":{"name":"go_info","help":"Information about the Go environment.","type":1,"metric":[{"label":[{"name":"version","value":"go1.9.1"}],"gauge":{"value":1}}]},"go_memstats_alloc_bytes":{"name":"go_memstats_alloc_bytes","help":"Number of bytes allocated and still in use.","type":1,"metric":[{"gauge":{"value":483504}}]}}`

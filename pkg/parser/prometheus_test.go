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
# HELP rpc_durations_histogram_seconds RPC latency distributions.
# TYPE rpc_durations_histogram_seconds histogram
rpc_durations_histogram_seconds_bucket{le="+Inf"} 0
rpc_durations_histogram_seconds_sum 0
rpc_durations_histogram_seconds_count 0
`

var promJSON = `{"go_gc_duration_seconds":{"name":"go_gc_duration_seconds","help":"A summary of the GC invocation durations.","type":"SUMMARY","metrics":[{"quantiles":{"0":"0","0.25":"0","0.5":"0","0.75":"0","1":"0"},"count":"0","sum":"0"}]},"go_goroutines":{"name":"go_goroutines","help":"Number of goroutines that currently exist.","type":"GAUGE","metrics":[{"value":"6"}]},"go_info":{"name":"go_info","help":"Information about the Go environment.","type":"GAUGE","metrics":[{"labels":{"version":"go1.9.1"},"value":"1"}]},"go_memstats_alloc_bytes":{"name":"go_memstats_alloc_bytes","help":"Number of bytes allocated and still in use.","type":"GAUGE","metrics":[{"value":"483504"}]},"rpc_durations_histogram_seconds":{"name":"rpc_durations_histogram_seconds","help":"RPC latency distributions.","type":"HISTOGRAM","metrics":[{"buckets":{"+Inf":"0"},"count":"0","sum":"0"}]}}`

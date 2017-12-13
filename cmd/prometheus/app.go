package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	normDomain            = flag.Float64("normal.domain", 0.0002, "The domain for the normal distribution.")
	normMean              = flag.Float64("normal.mean", 0.00001, "The mean for the normal distribution.")
	rpcDurationsHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "rpc_durations_histogram_seconds",
		Help:    "RPC latency distributions.",
		Buckets: prometheus.LinearBuckets(*normMean-5**normDomain, .5**normDomain, 20),
	})
)

func init() {
	prometheus.MustRegister(rpcDurationsHistogram)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("port is empty")
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(net.JoinHostPort("", port), nil))
}

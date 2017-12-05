package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("port is empty")
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(net.JoinHostPort("", port), nil))
}

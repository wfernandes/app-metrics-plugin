package main

import (
	"expvar"
	"log"
	"net"
	"net/http"
	"os"
)

// This is an app that will provide a variety of expvar style metrics
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("port is empty")
	}

	conn, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		log.Fatalf("unable to create listener: %s", err)
	}
	mInt := expvar.NewInt("metric.int")
	mInt.Set(10)

	mFloat := expvar.NewFloat("metric.float")
	mFloat.Set(123.345)

	mString := expvar.NewString("metric.string")
	mString.Set("expvarApp")

	mMap := expvar.NewMap("metric.map")
	mMap.Add("metric1", 10)
	mMap.Add("metric2", 11)

	http.Serve(conn, nil)
}

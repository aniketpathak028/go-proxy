package main

import (
	"flag"
	"log"
	"net/http"

	proxy "github.com/aniketpathak028/proxy-go/pkg/server"
)

func main() {
	port := flag.String("port", "8080", "Port to run the proxy on")
	flag.Parse()

	p := proxy.NewProxy()

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: p,
	}

	log.Printf("Starting proxy server on :%s", *port)
	log.Printf("Configure your browser to use http://localhost:%s as HTTP proxy", *port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

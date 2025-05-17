package proxy

import (
	"io"
	"log"
	"net/http"
)

type Proxy struct {
	client *http.Client
}

func NewProxy() *Proxy {
	return &Proxy{
		client: &http.Client{},
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// need separate logic to handle https requests
	if r.Method == http.MethodConnect {
		http.Error(w, "HTTPS not supported yet", http.StatusNotImplemented)
		return
	}

	log.Printf("Proxying request: %s %s", r.Method, r.URL.String())

	// fwd request to the destination server
	proxyReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// copy headers from original request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// send request to the destination server
	resp, err := p.client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// fwd response headers back to client
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// set status code
	w.WriteHeader(resp.StatusCode)

	// sets the response body
	io.Copy(w, resp.Body)

	log.Printf("Completed: %s %s - %d", r.Method, r.URL.String(), resp.StatusCode)
}

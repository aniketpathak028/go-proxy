package proxy

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aniketpathak028/proxy-go/pkg/cache"
	"github.com/aniketpathak028/proxy-go/pkg/utils"
)

// proxy struct
type Proxy struct {
	Client *http.Client
	Cache  cache.Cache
}

// constructor
func NewProxy() *Proxy {
	return &Proxy{
		Client: &http.Client{},
		Cache:  cache.NewMemCache(),
	}
}

// handler function for proxy
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check if the CONNECT method is requested by https
	if r.Method == http.MethodConnect {
		http.Error(w, "HTTPS not supported yet", http.StatusNotImplemented)
		return
	}

	log.Printf("Proxying request: %s %s", r.Method, r.URL.String())

	// check if request is cacheable
	if utils.IsCacheable(r) {
		cacheKey := utils.GetCacheKey(r)

		// check if a cached response exists
		if entry, found := p.Cache.Get(cacheKey); found {
			log.Printf("cache hit for %s", r.URL.String())

			// check if we can use cond req
			if entry.ETag != "" || entry.LastModified != "" {
				proxyReq, err := http.NewRequest(r.Method, r.URL.String(), nil)
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

				// add conditional headers
				if entry.ETag != "" {
					proxyReq.Header.Set("If-None-Match", entry.ETag)
				}
				if entry.LastModified != "" {
					proxyReq.Header.Set("If-Modified-Since", entry.LastModified)
				}

				resp, err := p.Client.Do(proxyReq)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer resp.Body.Close()

				// server returns 304 not modified, use cache
				if resp.StatusCode == http.StatusNotModified {
					log.Printf("using validated cache for: %s", r.URL.String())

					// forward cached headers
					for name, values := range entry.Response.Header {
						for _, value := range values {
							w.Header().Add(name, value)
						}
					}

					// set status code
					w.WriteHeader(entry.Response.StatusCode)

					// write cached body
					io.Copy(w, bytes.NewReader(entry.Body))
					return
				}

				// if got a new response, update cache
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// create a new resp and cache body
				newResp := *resp
				newResp.Body = io.NopCloser(bytes.NewReader(body))

				p.updateCache(cacheKey, &newResp, body)

				// fwd response headers
				for name, values := range resp.Header {
					for _, value := range values {
						w.Header().Add(name, value)
					}
				}

				// set status code
				w.WriteHeader(resp.StatusCode)

				// write body
				w.Write(body)
				return
			}

			// use cached resp directly
			for name, values := range entry.Response.Header {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}

			w.WriteHeader(entry.Response.StatusCode)
			io.Copy(w, bytes.NewReader(entry.Body))
			return
		}

	}

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
	resp, err := p.Client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// read the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create a new response with the body for caching
	newResp := *resp
	newResp.Body = io.NopCloser(bytes.NewReader(body))

	// cache the response if it's cacheable
	if utils.IsCacheable(r) && resp.StatusCode == http.StatusOK {
		cacheKey := utils.GetCacheKey(r)
		p.updateCache(cacheKey, &newResp, body)
	}

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

// stores a response in the cache
func (p *Proxy) updateCache(key string, resp *http.Response, body []byte) {
	// check cache control headers
	cc := resp.Header.Get("Cache-Control")
	if strings.Contains(cc, "no-store") || strings.Contains(cc, "private") {
		return
	}

	entry := &cache.CacheEntry{
		Response:     resp,
		Body:         body,
		CachedAt:     time.Now(),
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}

	// set expiration time based on cache-control or expires header
	if maxAge, ok := utils.ParseMaxAge(cc); ok {
		entry.ExpiresAt = time.Now().Add(maxAge)
	} else if expires := resp.Header.Get("Expires"); expires != "" {
		if expiresTime, err := time.Parse(time.RFC1123, expires); err == nil {
			entry.ExpiresAt = expiresTime
		}
	} else {
		// default cache time: 1 hour
		entry.ExpiresAt = time.Now().Add(1 * time.Hour)
	}

	p.Cache.Set(key, entry)
	log.Printf("Cached: %s (expires: %v)", key, entry.ExpiresAt)
}

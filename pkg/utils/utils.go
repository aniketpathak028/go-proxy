package utils

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// generates a unique string for an http request to store in cache map
func GetCacheKey(r *http.Request) string {
	return r.Method + ":" + r.URL.String()
}

// check if the request should be cached
func IsCacheable(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}

	// cache-control: no-store
	if cc := r.Header.Get("Cahce-Control"); strings.Contains(cc, "no-store") {
		return false
	}

	return true
}

// extract max-age from cache-control header
func ParseMaxAge(header string) (time.Duration, bool) {
	if header == "" {
		return 0, false
	}

	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "max-age=") {
			ageStr := strings.TrimPrefix(part, "max-age=")
			age, err := strconv.Atoi(ageStr)
			if err != nil {
				return 0, false
			}
			return time.Duration(age) * time.Second, true
		}
	}

	return 0, false
}

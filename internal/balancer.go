package lb

import (
	"log"
	configProvider "nanoLB/internal/config"
	"net/http"
)

var config = configProvider.GetConfig()

func LoadBalancer(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttempts(r)
	log.Println("Trying", attempts, "time for", r.RemoteAddr)
	if attempts > config.MaxAttempts {
		HttpHtmlError(w, ServiceUnavailable.String(), http.StatusServiceUnavailable)
		return
	}
	algo := GetAlgo(config.Algorithm)
	server := GetServerPool().next(algo)
	log.Println("Serving with", server.URL.String(), "to", r.RemoteAddr)
	if server != nil {
		server.ReverseProxy.ServeHTTP(w, r)
		return
	}
	HttpHtmlError(w, ServiceUnavailable.String(), http.StatusServiceUnavailable)
}

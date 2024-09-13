package lb

import (
	"log"
	configProvider "nanoLB/internal/config"
	"net/http"
)

var Config = configProvider.GetConfig()

func LoadBalancer(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttempts(r)
	log.Println("Trying", attempts, "time for", r.RemoteAddr)
	if attempts > Config.MaxAttempts {
		HttpHtmlError(w, ServiceUnavailable.String(), http.StatusServiceUnavailable)
		return
	}
	algo := GetAlgo(Config.Algorithm, GetServerPool())
	server := GetServerPool().next(algo)
	if server != nil {
		log.Println("Serving with", server.URL.String(), "to", r.RemoteAddr)
		server.ReverseProxy.ServeHTTP(w, r)
		return
	}
	HttpHtmlError(w, ServiceUnavailable.String(), http.StatusServiceUnavailable)
}

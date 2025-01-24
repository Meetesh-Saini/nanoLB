package lb

import (
	configProvider "nanoLB/internal/config"
	"nanoLB/internal/log"
	"net/http"
)

func LoadBalancer(w http.ResponseWriter, r *http.Request) {
	var Config = configProvider.GetConfig()
	attempts := GetAttempts(r)
	log.Logger.Infof("Trying %v time for %v", attempts, r.RemoteAddr)
	if attempts > Config.MaxAttempts {
		HttpHtmlError(w, ServiceUnavailable.String(), http.StatusServiceUnavailable)
		return
	}
	algo := GetAlgo(Config.Algorithm, GetServerPool())
	server := GetServerPool().next(algo)
	if server != nil {
		log.Logger.Infof("Serving with %v to %v", server.URL.String(), r.RemoteAddr)
		server.ReverseProxy.ServeHTTP(w, r)
		return
	}
	HttpHtmlError(w, ServiceUnavailable.String(), http.StatusServiceUnavailable)
}

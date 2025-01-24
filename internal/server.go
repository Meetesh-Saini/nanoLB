package lb

import (
	"context"
	"errors"
	"nanoLB/internal/config"
	"nanoLB/internal/log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Server struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	Healthy      bool
	Connections  uint
	Weight       int64
	mux          sync.RWMutex
}

type ServerPool struct {
	pool       []*Server
	poolLookup map[string]*Server
	mux        sync.RWMutex
}

var (
	serverPool     *ServerPool
	onceServerPool sync.Once
)

func GetServerPool() *ServerPool {
	onceServerPool.Do(func() {
		serverPool = &ServerPool{poolLookup: make(map[string]*Server)}
	})
	return serverPool
}

func GetServer(serverURL string, weight int64) *Server {
	serverUrl, err := url.Parse(serverURL)
	if err != nil {
		log.Logger.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(serverUrl)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Logger.Infof("[%s] %s", serverUrl.String(), e.Error())
		if errors.Is(e, context.Canceled) {
			return
		}
		retries := GetRetries(request)
		if retries < config.GetConfig().MaxRetries {
			<-time.After(10 * time.Millisecond)
			ctx := context.WithValue(request.Context(), Retry, retries+1)
			proxy.ServeHTTP(writer, request.WithContext(ctx))
			return
		}

		// after 3 retries, mark this backend as down
		serverPool.SetServerHealth(serverUrl.String(), false)

		// if the same request routing for few attempts with different backends, increase the count
		attempts := GetAttempts(request)
		log.Logger.Infof("%s(%s) Attempting retry %d", request.RemoteAddr, request.URL.Path, attempts)
		ctx := context.WithValue(request.Context(), Attempts, attempts+1)
		LoadBalancer(writer, request.WithContext(ctx))
	}

	log.Logger.Infof("Configured server: %s", serverUrl)
	return &Server{
		URL:          serverUrl,
		Healthy:      true,
		ReverseProxy: proxy,
		Connections:  0,
		Weight:       weight,
	}
}

func (s *ServerPool) Add(server *Server) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.pool = append(s.pool, server)
	s.poolLookup[server.URL.String()] = server
}

func (s *ServerPool) next(a Algorithm) *Server {
	return a.GetNext(s)
}

func (s *Server) IsHealthy() (healthy bool) {
	s.mux.RLock()
	healthy = s.Healthy
	s.mux.RUnlock()
	return
}

package lb

import (
	"context"
	"log"
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

func GetServer(serverURL string) *Server {
	serverUrl, err := url.Parse(serverURL)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(serverUrl)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
		retries := GetRetries(request)
		if retries < config.MaxRetries {
			select {
			case <-time.After(10 * time.Millisecond):
				ctx := context.WithValue(request.Context(), Retry, retries+1)
				proxy.ServeHTTP(writer, request.WithContext(ctx))
			}
			return
		}

		// after 3 retries, mark this backend as down
		serverPool.SetServerHealth(serverUrl.String(), false)

		// if the same request routing for few attempts with different backends, increase the count
		attempts := GetAttempts(request)
		log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
		ctx := context.WithValue(request.Context(), Attempts, attempts+1)
		LoadBalancer(writer, request.WithContext(ctx))
	}

	log.Printf("Configured server: %s\n", serverUrl)
	return &Server{
		URL:          serverUrl,
		Healthy:      true,
		ReverseProxy: proxy,
		Connections:  0,
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

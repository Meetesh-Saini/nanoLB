package lb

import (
	"context"
	"nanoLB/internal/log"
	"net"
	"sync"
	"time"
)

var healthChechMutex sync.Mutex

func (sp *ServerPool) SetServerHealth(serverURL string, status bool) {
	sp.poolLookup[serverURL].mux.Lock()
	defer sp.poolLookup[serverURL].mux.Unlock()
	sp.poolLookup[serverURL].Healthy = status
}

func (sp *ServerPool) HealthCheck() {
	var wg sync.WaitGroup
	ch := make(chan struct {
		server    *Server
		isHealthy bool
	}, len(sp.pool))

	// Limit concurrent health checks
	guard := make(chan struct{}, Config.MaxConcurrentHealthChecks)

	for _, s := range sp.pool {
		wg.Add(1)
		go func(server *Server) {
			defer wg.Done()

			// Acquire guard
			guard <- struct{}{}

			isHealthy := GetServerHealth(server)
			ch <- struct {
				server    *Server
				isHealthy bool
			}{server, isHealthy}

			// Release guard
			<-guard
		}(s)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Process results as they come in
	for result := range ch {
		sp.SetServerHealth(result.server.URL.String(), result.isHealthy)
		if result.isHealthy {
			log.Logger.Infof("%s [healthy]", result.server.URL.String())
		}
	}
}

func GetServerHealth(s *Server) bool {
	conn, err := net.DialTimeout("tcp", s.URL.Host, Config.HealthCheckTimeout)
	if err != nil {
		log.Logger.Infof("%s [dead] : %s", s.URL.String(), err)
		return false
	}
	defer conn.Close()
	return true
}

func HealthCheckRoutine(ctx context.Context) {
	t := time.NewTicker(Config.HealthCheckInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			healthChechMutex.Lock()
			log.Logger.Info("Starting health check...")

			// Run health check and log success or failure
			GetServerPool().HealthCheck()
			log.Logger.Info("Health check completed successfully")

			healthChechMutex.Unlock() // Release lock after the health check completes

		case <-ctx.Done():
			log.Logger.Info("Health check routine stopping")
			return
		}
	}
}

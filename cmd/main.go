package main

import (
	"context"
	"flag"
	"fmt"
	lb "nanoLB/internal"
	"nanoLB/internal/config"
	"nanoLB/internal/log"
	"net/http"
	"strings"
)

func main() {
	// Define flags
	var servers string
	var port int
	var configPath string

	flag.StringVar(&servers, "servers", "", "Comma-separated list of server URLs")
	flag.IntVar(&port, "port", 9696, "Port to listen on")
	flag.StringVar(&configPath, "config", "", "Path to configuration file")

	// Parse flags
	flag.Parse()

	// Print values for demonstration
	fmt.Println("Server URLs:", servers)
	fmt.Println("Port:", port)
	fmt.Println("Config Path:", configPath)

	lb.Config.Algorithm = config.WeightedRoundRobin

	if err := log.Init(); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}

	serverUrls := strings.Split(servers, ",")
	// TODO: Only for testing, remove afterwards
	w := []int64{5, 2, 10, 3, 7, 4}
	for c, url := range serverUrls {
		lb.GetServerPool().Add(lb.GetServer(url, w[c]))
		log.Logger.Infof("Added server to pool: %s", url)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb.LoadBalancer),
	}

	healthCheckCtx, healthCheckCtxCancel := context.WithCancel(context.Background())
	defer healthCheckCtxCancel()

	go lb.HealthCheckRoutine(healthCheckCtx)

	log.Logger.Infof("Load Balancer started at :%d", port)
	if err := server.ListenAndServe(); err != nil {
		log.Logger.Fatal(err)
	}
}

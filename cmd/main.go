package main

import (
	"context"
	"flag"
	"fmt"
	lb "nanoLB/internal"
	"nanoLB/internal/config"
	"nanoLB/internal/log"
	"net/http"
)

func main() {
	// Define flags
	var port int
	var configPath string

	flag.IntVar(&port, "port", 9696, "Port to listen on")
	flag.StringVar(&configPath, "config", "", "Path to configuration file")

	// Parse flags
	flag.Parse()

	// Use default config path if not provided
	if configPath == "" {
		configPath = "nanolb.toml"
		fmt.Println("Warning: Config path not provided, using default:", configPath)
	}

	// Print values for demonstration
	fmt.Println("Port:", port)
	fmt.Println("Config Path:", configPath)

	// Set config path
	config.SetConfigFilePath(configPath)

	// Load config
	config.GetConfig()

	// Set up logging
	if err := log.Init(); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}

	for _, server := range config.GetConfig().Servers {
		lb.GetServerPool().Add(lb.GetServer(server.URL, server.Weight))
		log.Logger.Infof("Added server to pool: %s", server.URL)
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

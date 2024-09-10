package main

import (
	"flag"
	"fmt"
	"log"
	lb "nanoLB/internal"
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

	serverUrls := strings.Split(servers, ",")
	for _, url := range serverUrls {
		lb.GetServerPool().Add(lb.GetServer(url))
		log.Printf("Added server to pool: %s\n", url)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb.LoadBalancer),
	}

	log.Printf("Load Balancer started at :%d\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

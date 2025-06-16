package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lokeshllkumar/load-balancer/internal/balancer"
	"github.com/lokeshllkumar/load-balancer/internal/config"
	"github.com/lokeshllkumar/load-balancer/internal/metrics"
	"github.com/lokeshllkumar/load-balancer/internal/proxy"
	"github.com/lokeshllkumar/load-balancer/internal/registry"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	metrics.InitMetrics()

	serviceRegistryClient, err := registry.NewServiceRegistryClient(cfg.ServiceRegistryType, cfg.ServiceRegsistryUrl)
	if err != nil {
		log.Fatalf("Failed to initialize service registry client: %v", err)
	}
	if grpcClient, ok := serviceRegistryClient.(*registry.GRPCRegistryClient); ok {
		defer grpcClient.Close()
	}

	backendManager := balancer.NewBackendManager(serviceRegistryClient, cfg.HealthCheckInterval, cfg.HealthCheckTimeout)
	go backendManager.StartBackendDiscovery(context.Background())
	go backendManager.StartHealthChecks(context.Background())

	var lbStrategy balancer.LoadBalancingStrategy
	switch cfg.Strategy {
	case "round_robin":
		lbStrategy = balancer.NewRoundRobinStrategy(backendManager)
	case "least_connections":
		lbStrategy = balancer.NewLeastConnectionsStrategy(backendManager)
	case "sticky_sessions":
		lbStrategy = balancer.NewStickySessionsStrategy(backendManager)
	default:
		log.Fatalf("Unsupported load balancing strategy: %s", cfg.Strategy)
	}

	reverseProxy := proxy.NewReverseProxyHandler(lbStrategy)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: metrics.PrometheusMiddleware(reverseProxy),
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Load balancer starting on :%d with strategy :%s", cfg.Port, cfg.Strategy)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<- stopChan
	log.Println("Shutting down load balancer gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	backendManager.Stop()

	log.Println("Load balancer shut down")
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lokeshllkumar/load-balancer/internal/backend"
	"github.com/lokeshllkumar/load-balancer/internal/balancer"
	"github.com/lokeshllkumar/load-balancer/internal/handler"
	_ "github.com/lokeshllkumar/load-balancer/internal/health"
	"github.com/lokeshllkumar/load-balancer/internal/metrics"
	"github.com/lokeshllkumar/load-balancer/internal/utils"
)

func main() {
	config, err := utils.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load backends: %v", err)
	}

	pool := backend.NewBackendPool()
	for _, b := range config.Backends {
		pool.AddBackend(b.URL, b.Weight, b.Sticky)
	}

	lb := balancer.NewLoadBalancer(pool, config.LoadBalancing.Strategy)


	// go health.StartHealthCheck(pool, config.Health.Interval)

	router := gin.Default()
	handler.RegisterRoutes(router, lb, config.LoadBalancing.Strategy)
	metrics.SetupMetrics(router)

	server := handler.CreateServer(router)

	go func() {
		log.Println("Starting load balancer on :8443...")
		if err := server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}

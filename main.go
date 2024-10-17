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
	"github.com/lokeshllkumar/load-balancer/internal/handler"
	"github.com/lokeshllkumar/load-balancer/internal/health"
	"github.com/lokeshllkumar/load-balancer/internal/metrics"
	"github.com/lokeshllkumar/load-balancer/internal/utils"
)

func main() {
	pool, err := utils.LoadBackendsFromConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load backends: %v", err)
	}

	go health.StartHealthCheck(pool, 10)

	router := gin.Default()
	handler.RegisterRoutes(router, pool)
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

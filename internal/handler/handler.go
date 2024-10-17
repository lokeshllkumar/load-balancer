package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lokeshllkumar/load-balancer/internal/backend"
	"github.com/lokeshllkumar/load-balancer/internal/balancer"
)

func RegisterRoutes(router *gin.Engine, pool *backend.BackendPool) {
	router.GET("/load-balancer", func(c *gin.Context) {
		backends := pool.GetAliveBackends()
		if len(backends) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No backends available"})
			return
		}
		backend := balancer.NewRoundRobin().GetNextBackendRR(backends)
		c.Redirect(http.StatusTemporaryRedirect, backend.URL)
	})
}

func CreateServer(router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:           ":8443",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

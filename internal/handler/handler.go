package handler

import (
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lokeshllkumar/load-balancer/internal/backend"
	"github.com/lokeshllkumar/load-balancer/internal/balancer"
)

func RegisterRoutes(router *gin.Engine, lb *balancer.LoadBalancer, strategy string) {
	router.GET("/load-balancer", func(c *gin.Context) {
		backends := lb.Pool.GetAliveBackends()
		if len(backends) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No backends available"})
			return
		}
		
		var selectedBackend *backend.Backend
		clientIP := getClientIP(c.Request)

		switch strategy {
		case "round-robin":
			selectedBackend = balancer.GetNextBackendRR(backends)
		case "least-connections":
			selectedBackend = balancer.GetNextBackendLC(backends)
		case "sticky":
			selectedBackend = lb.GetBackendSS(clientIP)
		}
		c.Redirect(http.StatusTemporaryRedirect, selectedBackend.URL)
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

func getClientIP(req *http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return ip
}
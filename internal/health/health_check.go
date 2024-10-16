package health

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lokeshllkumar/load-balancer/internal/backend"
)

func StartHealthCheck(pool *backend.BackendPool, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for range ticker.C {
		backends := pool.GetAliveBackends()
		var wg sync.WaitGroup

		for _, b := range backends {
			wg.Add(1)
			go func(backend *backend.Backend) {
				defer wg.Done()
				checkBackend(backend)
			}(b)
		}

		wg.Wait()
	}
}

func checkBackend(b *backend.Backend) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	res, err := client.Get(b.URL + "/health")
	if err != nil || res.StatusCode != http.StatusOK {
		b.Alive = false
		log.Printf("[HealthCheck] Backend %s is Down\n", b.URL)
		return
	}

	b.Alive = true
	log.Printf("[HealthCheck] Backend %s is Up", b.URL)
}

func RegisterHealthRoute(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "Load balancer is healthy",
		})
	})
}

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lokeshllkumar/load-balancer/internal/backend"
)

func RegisterAPIRoutes(router *gin.Engine, pool *backend.BackendPool) {
	router.POST("/api/backends", func(c *gin.Context) {
		addBackend(c, pool)
	})

	router.DELETE("/api/backends", func(c *gin.Context) {
		removeBackend(c, pool)
	})

	router.GET("/api/backends", func(c *gin.Context) {
		listBackends(c, pool)
	})

	router.POST("/api/backends/increment", func(c *gin.Context) {
		incrementConnections(c, pool)
	})

	router.POST("/api/backends/decrement", func(c *gin.Context) {
		decrementConnections(c, pool)
	})
}

func addBackend(c *gin.Context, pool *backend.BackendPool) {
	var req struct {
		URL    string `json:"url" binding:"required"`
		Weight int    `json:"weight" binding:"required"`
		Sticky bool   `json:"sticky"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool.AddBackend(req.URL, req.Weight, req.Sticky)
	c.JSON(http.StatusOK, gin.H{"status": "Backend added", "url": req.URL})
}

func removeBackend(c *gin.Context, pool *backend.BackendPool) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := pool.RemoveBackend(req.URL); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backend not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Backend removed", "url": req.URL,
	})
}

func listBackends(c *gin.Context, pool *backend.BackendPool) {
	backends := pool.GetAliveBackends()

	var backendStatuses []gin.H
	for _, b := range pool.GetBackends() {
		backendStatuses = append(backendStatuses, gin.H{
			"url":         b.URL,
			"alive":       b.Alive,
			"weight":      b.Weight,
			"sticky":      b.Sticky,
			"connections": b.Connections,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"backends": backendStatuses,
		"alive":    len(backends),
		"total":    len(pool.GetBackends()),
	})
}

func incrementConnections(c *gin.Context, pool *backend.BackendPool) {

}

func decrementConnections(c *gin.Context, pool *backend.BackendPool) {
	
}
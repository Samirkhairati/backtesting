package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samirkhairati/backtesting/core/api/middleware"
)

// SetupRouter initializes and configures the Gin router with all endpoints
func SetupRouter() *gin.Engine {
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	return r
}

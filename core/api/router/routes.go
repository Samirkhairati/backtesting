package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samirkhairati/backtesting/core/api/middleware"
)

// SetupRouter initializes and configures the Gin router with all endpoints
func SetupRouter() *gin.Engine {
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	return r
}

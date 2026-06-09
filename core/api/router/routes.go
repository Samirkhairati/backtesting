package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin router with all endpoints
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	return r
}

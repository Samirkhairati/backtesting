package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samirkhairati/backtesting/core/api/handlers"
	"github.com/samirkhairati/backtesting/core/api/middleware"
)

// SetupRouter initializes and configures the Gin router with all endpoints
func SetupRouter(db *pgxpool.Pool) *gin.Engine {
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

	// Handlers Initialization
	authHandler := &handlers.AuthHandler{DB: db}

	// API Group Setup
	v1 := r.Group("/api/v1")

	// Unsecured routes
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// Secured routes (Requires JWT)
	// protected := v1.Group("/")
	// protected.Use(middleware.RequireAuth())
	// {
	//     // Future protected routes will go here (e.g. protected.POST("/backtests"))
	// }

	return r
}

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

	authHandler := &handlers.AuthHandler{DB: db}
	experimentHandler := &handlers.ExperimentHandler{DB: db}

	// v1 API Routes
	v1 := r.Group("/api/v1")
	{
		// Public Routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected Routes
		experiments := v1.Group("/experiments")
		experiments.Use(middleware.RequireAuth())
		{
			experiments.POST("", experimentHandler.CreateExperiment)
			experiments.GET("", experimentHandler.GetExperiments)
			experiments.GET("/:id", experimentHandler.GetExperimentByID)
		}
	}

	return r
}

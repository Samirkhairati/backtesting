package router

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samirkhairati/backtesting/core/api/handlers"
	"github.com/samirkhairati/backtesting/core/api/middleware"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
)

// SetupRouter initializes and configures the Gin router with all endpoints
func SetupRouter(db *pgxpool.Pool, coordinatorClient pb.CoordinatorServiceClient, bus *handlers.EventBus) *gin.Engine {
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	// Configure CORS
	// The bundled client talks to the API on its own origin via a reverse proxy,
	// so CORS is only needed when the API is hosted on a different origin.
	// CORS_ALLOWED_ORIGINS is a comma-separated list; in debug mode (and only
	// there) an empty value allows every origin for convenience.
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	if origins := splitList(os.Getenv("CORS_ALLOWED_ORIGINS")); len(origins) > 0 {
		corsConfig.AllowOrigins = origins
		r.Use(cors.New(corsConfig))
	} else if gin.Mode() != gin.ReleaseMode {
		corsConfig.AllowAllOrigins = true
		r.Use(cors.New(corsConfig))
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	authHandler := &handlers.AuthHandler{DB: db}
	experimentHandler := &handlers.ExperimentHandler{
		DB:                db,
		CoordinatorClient: coordinatorClient,
	}
	dataHandler := &handlers.DataHandler{DB: db}
	eventHandler := &handlers.EventHandler{Bus: bus}

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
			experiments.POST("/:id/cancel", experimentHandler.CancelExperiment)
			experiments.GET("/:id/events", eventHandler.StreamEvents)
		}

		data := v1.Group("/data")
		data.Use(middleware.RequireAuth())
		{
			data.GET("/symbols", dataHandler.GetSymbols)
			data.GET("/symbols/search", dataHandler.SearchSymbols)
		}
	}

	return r
}

// splitList parses a comma-separated env value, dropping blanks.
func splitList(v string) []string {
	var out []string
	for _, item := range strings.Split(v, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

package handlers

import (
	"context"
	"net/http"
	"time"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
	"go.uber.org/zap"
)

type ExperimentHandler struct {
	DB                *pgxpool.Pool
	CoordinatorClient pb.CoordinatorServiceClient
}

// CreateExperiment creates a new pending experiment
func (h *ExperimentHandler) CreateExperiment(c *gin.Context) {
	// Extract the userID that was injected by the RequireAuth middleware
	userID := c.MustGet("user_id").(string)

	var req models.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, tf := range req.Timeframes {
		start, err := time.Parse(time.RFC3339, tf.Start)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		end, err := time.Parse(time.RFC3339, tf.End)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		if start.After(end) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start date must be before end date"})
			return
		}
		if end.Sub(start).Hours() < 24*28 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Timeframe must be at least 1 month long (28 days)"})
			return
		}
	}

	configJSONBytes, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to marshal experiment config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process experiment configuration"})
		return
	}

	grpcReq := &pb.CreateExperimentRequest{
		UserId:      userID,
		Name:        req.Name,
		Description: req.Description,
		ConfigJson:  string(configJSONBytes),
	}

	resp, err := h.CoordinatorClient.CreateExperiment(c.Request.Context(), grpcReq)
	if err != nil {
		logger.Log.Error("Failed to create experiment via Coordinator", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create experiment in coordinator"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Experiment created successfully",
		"experiment_id": resp.ExperimentId,
		"status":        resp.Status,
	})
}

// GetExperiments returns all experiments for the authenticated user
func (h *ExperimentHandler) GetExperiments(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	rows, err := h.DB.Query(
		context.Background(),
		"SELECT id, user_id, name, description, config, status, created_at FROM experiments WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		logger.Log.Error("Failed to query experiments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch experiments"})
		return
	}
	defer rows.Close()

	var experiments []models.Experiment
	for rows.Next() {
		var exp models.Experiment
		if err := rows.Scan(
			&exp.ID,
			&exp.UserID,
			&exp.Name,
			&exp.Description,
			&exp.Config,
			&exp.Status,
			&exp.CreatedAt,
		); err != nil {
			logger.Log.Error("Failed to scan experiment", zap.Error(err))
			continue
		}
		experiments = append(experiments, exp)
	}

	if experiments == nil {
		experiments = []models.Experiment{}
	}

	c.JSON(http.StatusOK, gin.H{"experiments": experiments})
}

// GetExperimentByID returns a specific experiment if it belongs to the authenticated user
func (h *ExperimentHandler) GetExperimentByID(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	expID := c.Param("id")

	var exp models.Experiment
	err := h.DB.QueryRow(
		context.Background(),
		"SELECT id, user_id, name, description, config, status, created_at FROM experiments WHERE id = $1 AND user_id = $2",
		expID,
		userID,
	).Scan(
		&exp.ID,
		&exp.UserID,
		&exp.Name,
		&exp.Description,
		&exp.Config,
		&exp.Status,
		&exp.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experiment not found"})
		return
	}

	c.JSON(http.StatusOK, exp)
}

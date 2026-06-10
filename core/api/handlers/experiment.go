package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"go.uber.org/zap"
)

type ExperimentHandler struct {
	DB *pgxpool.Pool
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

	var experimentID string
	err := h.DB.QueryRow(
		context.Background(),
		"INSERT INTO experiments (user_id, name, description, config, status) VALUES ($1, $2, $3, $4, 'PENDING') RETURNING id",
		userID,
		req.Name,
		req.Description,
		req,
	).Scan(&experimentID)

	if err != nil {
		logger.Log.Error("Failed to insert experiment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create experiment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Experiment created successfully",
		"experiment_id": experimentID,
		"status":        "PENDING",
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

	c.JSON(http.StatusOK, experiments)
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

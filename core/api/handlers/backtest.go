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

type BacktestHandler struct {
	DB *pgxpool.Pool
}

// CreateBacktest creates a new pending backtest job
func (h *BacktestHandler) CreateBacktest(c *gin.Context) {
	// Extract the userID that was injected by the RequireAuth middleware
	userID := c.MustGet("user_id").(string)

	var req models.CreateBacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var jobID string
	err := h.DB.QueryRow(
		context.Background(),
		"INSERT INTO backtest_jobs (user_id, strategy_name, symbols, status) VALUES ($1, $2, $3, 'PENDING') RETURNING id",
		userID,
		req.StrategyName,
		req.Symbols,
	).Scan(&jobID)

	if err != nil {
		logger.Log.Error("Failed to insert backtest job", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create backtest job"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Backtest job created successfully",
		"job_id":  jobID,
		"status":  "PENDING",
	})
}

// GetBacktests returns all backtest jobs for the authenticated user
func (h *BacktestHandler) GetBacktests(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	rows, err := h.DB.Query(
		context.Background(),
		"SELECT id, user_id, strategy_name, symbols, status, created_at, completed_at, result_metrics FROM backtest_jobs WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		logger.Log.Error("Failed to query backtest jobs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch backtests"})
		return
	}
	defer rows.Close()

	var jobs []models.BacktestJob
	for rows.Next() {
		var job models.BacktestJob
		if err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.StrategyName,
			&job.Symbols,
			&job.Status,
			&job.CreatedAt,
			&job.CompletedAt,
			&job.ResultMetrics,
		); err != nil {
			logger.Log.Error("Failed to scan backtest job", zap.Error(err))
			continue
		}
		jobs = append(jobs, job)
	}

	// If jobs is nil, return an empty array instead of null
	if jobs == nil {
		jobs = []models.BacktestJob{}
	}

	c.JSON(http.StatusOK, jobs)
}

// GetBacktestByID returns a specific backtest job if it belongs to the authenticated user
func (h *BacktestHandler) GetBacktestByID(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	jobID := c.Param("id")

	var job models.BacktestJob
	err := h.DB.QueryRow(
		context.Background(),
		"SELECT id, user_id, strategy_name, symbols, status, created_at, completed_at, result_metrics FROM backtest_jobs WHERE id = $1 AND user_id = $2",
		jobID,
		userID,
	).Scan(
		&job.ID,
		&job.UserID,
		&job.StrategyName,
		&job.Symbols,
		&job.Status,
		&job.CreatedAt,
		&job.CompletedAt,
		&job.ResultMetrics,
	)

	if err != nil {
		// If no rows are returned or ID format is invalid, return 404
		c.JSON(http.StatusNotFound, gin.H{"error": "Backtest job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

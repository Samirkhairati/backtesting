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

	ids := make([]string, 0, len(experiments))
	for _, e := range experiments {
		ids = append(ids, e.ID)
	}
	progress := h.loadProgress(c.Request.Context(), ids)
	now := time.Now().UTC()
	for i := range experiments {
		experiments[i].Progress = progress[experiments[i].ID]
		experiments[i].ServerTime = &now
	}

	c.JSON(http.StatusOK, gin.H{"experiments": experiments, "server_time": now})
}


// loadProgress returns job counts for the given experiments straight from the database.
// Counting here (instead of relying on live events, which have no replay) means a client
// can draw correct progress bars on its very first request.
func (h *ExperimentHandler) loadProgress(ctx context.Context, ids []string) map[string]*models.ExperimentProgress {
	out := make(map[string]*models.ExperimentProgress, len(ids))
	for _, id := range ids {
		out[id] = &models.ExperimentProgress{}
	}
	if len(ids) == 0 {
		return out
	}

	fill := func(table string, set func(p *models.ExperimentProgress, total, done int)) {
		rows, err := h.DB.Query(
			ctx,
			"SELECT experiment_id::text, COUNT(*), COUNT(*) FILTER (WHERE status IN ('COMPLETED','FAILED')) FROM "+table+" WHERE experiment_id = ANY($1::uuid[]) GROUP BY experiment_id",
			ids,
		)
		if err != nil {
			logger.Log.Error("Failed to load job progress", zap.Error(err), zap.String("table", table))
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			var total, done int
			if err := rows.Scan(&id, &total, &done); err == nil {
				if p, ok := out[id]; ok {
					set(p, total, done)
				}
			}
		}
	}
	fill("ingestion_jobs", func(p *models.ExperimentProgress, total, done int) {
		p.Ingestion = models.JobProgress{Total: total, Completed: done}
	})
	fill("backtest_jobs", func(p *models.ExperimentProgress, total, done int) {
		p.Backtesting = models.JobProgress{Total: total, Completed: done}
	})
	return out
}

// GetExperimentByID returns a specific experiment if it belongs to the authenticated user
func (h *ExperimentHandler) GetExperimentByID(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	expID := c.Param("id")

	var exp models.Experiment
	err := h.DB.QueryRow(
		context.Background(),
		"SELECT id, user_id, name, description, config, status, created_at, result_summary, equity_curves FROM experiments WHERE id = $1 AND user_id = $2",
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
		&exp.ResultSummary,
		&exp.EquityCurves,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experiment not found"})
		return
	}

	// Fetch status logs
	logRows, err := h.DB.Query(
		context.Background(),
		"SELECT status, created_at FROM experiment_status_logs WHERE experiment_id = $1 ORDER BY created_at ASC",
		expID,
	)
	if err == nil {
		defer logRows.Close()
		var logs []models.ExperimentStatusLog
		for logRows.Next() {
			var log models.ExperimentStatusLog
			if err := logRows.Scan(&log.Status, &log.CreatedAt); err == nil {
				logs = append(logs, log)
			}
		}
		exp.StatusLogs = logs
	} else {
		logger.Log.Error("Failed to query status logs", zap.Error(err))
	}

	exp.Progress = h.loadProgress(c.Request.Context(), []string{exp.ID})[exp.ID]
	now := time.Now().UTC()
	exp.ServerTime = &now

	c.JSON(http.StatusOK, exp)
}

// CancelExperiment cancels a pending or running experiment
func (h *ExperimentHandler) CancelExperiment(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	expID := c.Param("id")

	// Make sure the experiment belongs to the user
	var exists bool
	err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM experiments WHERE id = $1 AND user_id = $2)", expID, userID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experiment not found"})
		return
	}

	grpcReq := &pb.CancelExperimentRequest{
		ExperimentId: expID,
		UserId:       userID,
	}

	_, err = h.CoordinatorClient.CancelExperiment(c.Request.Context(), grpcReq)
	if err != nil {
		logger.Log.Error("Failed to cancel experiment via Coordinator", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel experiment in coordinator"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Experiment cancelled successfully",
	})
}

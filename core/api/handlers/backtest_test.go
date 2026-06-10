package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samirkhairati/backtesting/core/api/router"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"github.com/stretchr/testify/assert"
)

func TestBacktestEndpoints(t *testing.T) {
	logger.Init()
	db := SetupTestDB(t)
	defer db.Close()

	r := router.SetupRouter(db)
	
	token, _ := CreateTestUserAndLogin(t, r)



	var createdJobID string

	t.Run("Create Backtest", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.CreateBacktestRequest{
			StrategyName: "MACD",
			Symbols:      []string{"AAPL"},
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/backtests", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		createdJobID = response["job_id"].(string)
		assert.NotEmpty(t, createdJobID)
		assert.Equal(t, "PENDING", response["status"])
	})

	t.Run("Get Backtests List", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/backtests", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var jobs []models.BacktestJob
		json.Unmarshal(w.Body.Bytes(), &jobs)

		assert.GreaterOrEqual(t, len(jobs), 1)
		assert.Equal(t, createdJobID, jobs[0].ID)
	})

	t.Run("Get Specific Backtest", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/backtests/"+createdJobID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var job models.BacktestJob
		json.Unmarshal(w.Body.Bytes(), &job)

		assert.Equal(t, createdJobID, job.ID)
		assert.Equal(t, "MACD", job.StrategyName)
	})

	t.Run("Unauthorized Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/backtests", nil)
		// Missing Auth Header
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

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

func TestExperimentEndpoints(t *testing.T) {
	logger.Init()
	db := SetupTestDB(t)
	defer db.Close()

	r := router.SetupRouter(db)

	token, _ := CreateTestUserAndLogin(t, r)

	var experimentID string

	t.Run("Create_Experiment", func(t *testing.T) {
		reqBody := models.CreateExperimentRequest{
			Name:        "Test Experiment",
			Description: "A test experiment",
			Symbols:     []string{"AAPL", "MSFT"},
			Strategies: []models.StrategySweep{
				{
					Name: "SMA_Cross",
					Parameters: map[string][]interface{}{
						"fast": {10, 20},
						"slow": {50, 100},
					},
				},
			},
			Timeframes: []models.Timeframe{
				{Start: "2020-01-01", End: "2020-12-31"},
			},
		}

		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/experiments", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		experimentID = resp["experiment_id"].(string)
		assert.NotEmpty(t, experimentID)
		assert.Equal(t, "PENDING", resp["status"])
	})

	t.Run("Get_Experiments_List", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/experiments", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var exps []models.Experiment
		json.Unmarshal(w.Body.Bytes(), &exps)

		assert.Len(t, exps, 1)
		assert.Equal(t, experimentID, exps[0].ID)
		assert.Equal(t, "Test Experiment", exps[0].Name)
	})

	t.Run("Get_Specific_Experiment", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/experiments/"+experimentID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var exp models.Experiment
		json.Unmarshal(w.Body.Bytes(), &exp)

		assert.Equal(t, experimentID, exp.ID)
		assert.Equal(t, "Test Experiment", exp.Name)
	})

	t.Run("Unauthorized_Request", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/experiments", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

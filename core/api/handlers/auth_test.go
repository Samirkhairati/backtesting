package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/samirkhairati/backtesting/core/api/router"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthEndpoints(t *testing.T) {
	logger.Init()
	db := SetupTestDB(t)
	defer db.Close()

	r := router.SetupRouter(db)

	// Generate random username to avoid conflicts
	testUsername := "testuser_" + uuid.New().String()
	testPassword := "secret123"

	t.Run("Register User", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.RegisterRequest{
			Username: testUsername,
			Password: testPassword,
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["user_id"])
	})

	t.Run("Login User", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.LoginRequest{
			Username: testUsername,
			Password: testPassword,
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.AuthResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response.Token)
	})
}

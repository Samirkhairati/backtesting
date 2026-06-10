package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samirkhairati/backtesting/core/models"
	"github.com/stretchr/testify/require"
)

// SetupTestDB initializes and returns a Postgres connection pool for integration tests.
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/backtesting?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dbURL)
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	require.NoError(t, err)

	err = pool.Ping(context.Background())
	if err != nil {
		t.Skipf("Skipping integration test: database not available at %s", dbURL)
	}

	return pool
}

// CreateTestUserAndLogin registers a random new user, logs them in, and returns the JWT token.
func CreateTestUserAndLogin(t *testing.T, r *gin.Engine) (string, string) {
	testUsername := "testuser_" + uuid.New().String()
	testPassword := "secret123"

	// 1. Register
	regBody, _ := json.Marshal(models.RegisterRequest{Username: testUsername, Password: testPassword})
	req1, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(regBody))
	req1.Header.Set("Content-Type", "application/json")
	
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusCreated, w1.Code, "Failed to register test user")

	// 2. Login
	loginBody, _ := json.Marshal(models.LoginRequest{Username: testUsername, Password: testPassword})
	req2, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req2.Header.Set("Content-Type", "application/json")
	
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusOK, w2.Code, "Failed to login test user")

	var resp models.AuthResponse
	err := json.Unmarshal(w2.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Token)

	return resp.Token, testUsername
}

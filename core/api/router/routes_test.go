package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckRoute(t *testing.T) {
	// Setup the router
	router := SetupRouter()

	// Create a new HTTP GET request to /health
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	// Serve the HTTP request
	router.ServeHTTP(w, req)

	// Assert HTTP Status Code is 200 OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Assert Response Body
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

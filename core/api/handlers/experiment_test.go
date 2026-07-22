package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/samirkhairati/backtesting/core/api/router"
	"github.com/samirkhairati/backtesting/core/coordinator"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var mockExpID = uuid.New().String()

type mockCoordinatorClient struct {
	pb.CoordinatorServiceClient
}

func (m *mockCoordinatorClient) CreateExperiment(ctx context.Context, in *pb.CreateExperimentRequest, opts ...grpc.CallOption) (*pb.CreateExperimentResponse, error) {
	return &pb.CreateExperimentResponse{
		ExperimentId: mockExpID,
		Status:       "PENDING",
	}, nil
}

// --------------------------------------------------------------------------------
// 1. COMPONENT TEST: Uses Mock Coordinator Client
// --------------------------------------------------------------------------------
func TestExperimentEndpoints_Mocked(t *testing.T) {
	logger.Init()
	db := SetupTestDB(t)
	defer db.Close()

	mockClient := &mockCoordinatorClient{}
	r := router.SetupRouter(db, mockClient, nil)

	token, userID := CreateTestUserAndLogin(t, r)

	var experimentID string

	t.Run("Create_Experiment", func(t *testing.T) {
		reqBody := models.CreateExperimentRequest{
			Name:        "Test Experiment Mocked",
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
				{Start: "2020-01-01T00:00:00Z", End: "2020-12-31T00:00:00Z"},
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
		assert.Equal(t, mockExpID, experimentID)
		assert.Equal(t, "PENDING", resp["status"])
	})

	// Manually insert into DB since our Mock didn't do it!
	_, err := db.Exec(context.Background(),
		"INSERT INTO experiments (id, user_id, name, description, config, status) VALUES ($1, $2, $3, $4, $5, $6)",
		experimentID, userID, "Test Experiment Mocked", "A test experiment", "{}", "PENDING",
	)
	require.NoError(t, err)

	t.Run("Get_Experiments_List", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/experiments", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Experiments []models.Experiment `json:"experiments"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		exps := resp.Experiments

		require.Len(t, exps, 1)
		assert.Equal(t, experimentID, exps[0].ID)
		assert.Equal(t, userID, exps[0].UserID)
		assert.Equal(t, "Test Experiment Mocked", exps[0].Name)
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
		assert.Equal(t, userID, exp.UserID)
		assert.Equal(t, "Test Experiment Mocked", exp.Name)
	})
}

// --------------------------------------------------------------------------------
// 2. END-TO-END INTEGRATION TEST: Spins up REAL Coordinator Server
// --------------------------------------------------------------------------------
func TestExperimentEndpoints_Integration(t *testing.T) {
	logger.Init()
	db := SetupTestDB(t)
	defer db.Close()

	// 1. Spin up the real Coordinator gRPC Server on a random local port
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	coordServer := coordinator.NewServer(db, nil)
	pb.RegisterCoordinatorServiceServer(grpcServer, coordServer)

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	// 2. Create a real gRPC Client connecting to it
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	realClient := pb.NewCoordinatorServiceClient(conn)

	// 3. Mount the API with the REAL client
	r := router.SetupRouter(db, realClient, nil)

	token, userID := CreateTestUserAndLogin(t, r)
	var experimentID string

	t.Run("E2E_Create_Experiment", func(t *testing.T) {
		reqBody := models.CreateExperimentRequest{
			Name:        "E2E Integration Experiment",
			Description: "Hits the real coordinator",
			Symbols:     []string{"TSLA_TEST_UNIQUE"},
			Strategies: []models.StrategySweep{
				{
					Name: "EMA_Cross",
					Parameters: map[string][]interface{}{
						"fast": {10},
						"slow": {50},
					},
				},
			},
			Timeframes: []models.Timeframe{
				{Start: "2021-01-01T00:00:00Z", End: "2021-12-31T00:00:00Z"},
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
		assert.Equal(t, "PENDING_INGESTION", resp["status"])
	})

	// NO MANUAL INSERT NEEDED! The real coordinator wrote it to the database!

	t.Run("E2E_Get_Experiments_List", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/experiments", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Experiments []models.Experiment `json:"experiments"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		exps := resp.Experiments

		require.Len(t, exps, 1)
		assert.Equal(t, experimentID, exps[0].ID)
		assert.Equal(t, userID, exps[0].UserID)
		assert.Equal(t, "E2E Integration Experiment", exps[0].Name)
	})
}

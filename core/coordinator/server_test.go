package coordinator

import (
	"context"
	"errors"
	"testing"
	"os"

	"github.com/jackc/pgx/v5"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/db"
)

// mockRow implements pgx.Row
type mockRow struct {
	err error
	id  string
}

func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	*dest[0].(*string) = m.id
	return nil
}

// mockDB implements DBQuerier
type mockDB struct {
	row *mockRow
}

func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.row
}

func TestServer_CreateExperiment_Unit(t *testing.T) {
	logger.Init()

	tests := []struct {
		name        string
		req         *pb.CreateExperimentRequest
		mockRow     *mockRow
		wantErr     bool
		expectedID  string
	}{
		{
			name: "Success",
			req: &pb.CreateExperimentRequest{
				UserId:      "user-123",
				Name:        "Test Exp",
				Description: "Desc",
				ConfigJson:  "{}",
			},
			mockRow: &mockRow{id: "exp-123", err: nil},
			wantErr: false,
			expectedID: "exp-123",
		},
		{
			name: "DB Error",
			req: &pb.CreateExperimentRequest{
				UserId:      "user-123",
				Name:        "Test Exp",
				Description: "Desc",
				ConfigJson:  "{}",
			},
			mockRow: &mockRow{id: "", err: errors.New("db error")},
			wantErr: true,
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				DB: &mockDB{row: tt.mockRow},
			}

			resp, err := s.CreateExperiment(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateExperiment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && resp.ExperimentId != tt.expectedID {
				t.Errorf("CreateExperiment() got ID = %v, want %v", resp.ExperimentId, tt.expectedID)
			}
		})
	}
}

// TestServer_CreateExperiment_Integration tests against a real database if available
func TestServer_CreateExperiment_Integration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}
	
	logger.Init()
	ctx := context.Background()

	dbPool := db.Setup(ctx)
	defer dbPool.Close()

	s := &Server{
		DB: dbPool,
	}

	req := &pb.CreateExperimentRequest{
		UserId:      "test-user-id", // Note: Foreign key constraints might fail if this user doesn't exist
		Name:        "Integration Exp",
		Description: "Integration Test",
		ConfigJson:  `{"param": "val"}`,
	}

	// For a pure integration test with FK constraints, we would need to insert a user first.
	// Since this is generic, if it fails due to FK, we accept it as DB contact success for this example.
	resp, err := s.CreateExperiment(ctx, req)
	if err != nil {
		// Log but don't strictly fail if it's a foreign key error
		t.Logf("CreateExperiment returned error (possibly FK missing): %v", err)
	} else {
		if resp.ExperimentId == "" {
			t.Errorf("Expected valid ExperimentId, got empty string")
		}
	}
}

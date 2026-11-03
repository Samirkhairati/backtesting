package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// A worker heartbeats every 5s; allow a few missed beats before counting it as gone.
const workerLiveWindow = "15 seconds"

type WorkerHandler struct {
	DB *pgxpool.Pool
}

// GetWorkers returns how many workers of each kind are currently alive.
func (h *WorkerHandler) GetWorkers(c *gin.Context) {
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT kind, COUNT(*)
		FROM workers
		WHERE status = 'ACTIVE' AND last_heartbeat > CURRENT_TIMESTAMP - $1::interval
		GROUP BY kind
	`, workerLiveWindow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workers"})
		return
	}
	defer rows.Close()

	counts := map[string]int{"ingestion": 0, "backtesting": 0}
	for rows.Next() {
		var kind string
		var n int
		if err := rows.Scan(&kind, &n); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse workers"})
			return
		}
		counts[kind] = n
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error during worker iteration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ingestion": counts["ingestion"], "backtesting": counts["backtesting"]})
}

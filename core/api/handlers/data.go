package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DataHandler struct {
	DB *pgxpool.Pool
}

func (h *DataHandler) GetSymbols(c *gin.Context) {
	rows, err := h.DB.Query(context.Background(), "SELECT DISTINCT symbol FROM daily_bars ORDER BY symbol ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch symbols from database"})
		return
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse symbols"})
			return
		}
		symbols = append(symbols, symbol)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error during symbol iteration"})
		return
	}

	// If no symbols found, return empty array instead of null
	if symbols == nil {
		symbols = []string{}
	}

	c.JSON(http.StatusOK, gin.H{"symbols": symbols})
}

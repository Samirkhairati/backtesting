package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

func (h *DataHandler) SearchSymbols(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, gin.H{"symbols": []map[string]string{}})
		return
	}

	searchURL := fmt.Sprintf("https://query2.finance.yahoo.com/v1/finance/search?q=%s", url.QueryEscape(query))
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", searchURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create search request"})
		return
	}

	// Add user-agent as yahoo blocks some generic ones
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to proxy search request"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Upstream search API error"})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read upstream response"})
		return
	}

	var yahooResp struct {
		Quotes []struct {
			Symbol    string `json:"symbol"`
			ShortName string `json:"shortname"`
			LongName  string `json:"longname"`
			QuoteType string `json:"quoteType"`
		} `json:"quotes"`
	}

	if err := json.Unmarshal(body, &yahooResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse upstream response"})
		return
	}

	var results []map[string]string
	for _, q := range yahooResp.Quotes {
		if q.QuoteType == "EQUITY" || q.QuoteType == "ETF" {
			name := q.ShortName
			if name == "" {
				name = q.LongName
			}
			if name == "" {
				name = q.Symbol
			}
			results = append(results, map[string]string{
				"symbol": q.Symbol,
				"name":   name,
			})
		}
	}

	if results == nil {
		results = []map[string]string{}
	}

	c.JSON(http.StatusOK, gin.H{"symbols": results})
}

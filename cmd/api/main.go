package main

import (
	"log"

	"github.com/samirkhairati/backtesting/core/api/router"
)

func main() {
	log.Println("Starting API Server on port 8080...")

	r := router.SetupRouter()

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}

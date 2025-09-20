package main

import (
	"log"
	"os"

	"voicegenie/internal/api"
	"voicegenie/internal/config"
	"voicegenie/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file (if exists)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Initialize logger
	logger.Init(cfg.Log.Level, cfg.Log.Format)

	// Create and start server
	server := api.NewServer(cfg)
	if err := server.Start(); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
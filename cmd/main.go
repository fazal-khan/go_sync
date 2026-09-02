package main

import (
	"log"
	"log/slog"

	"github.com/fazal-khan/go_sync/internal/logger"
	"github.com/joho/godotenv"
)

func main() {
	// initialize the application
	log.Println("in main")
	// load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading .env file", err)
	}

	log.Println("environment variables loaded successfully")

	initLogger()
	slog.Info("Logger initialized successfully")
}

func initLogger() {
	logger.NewConsoleLogger()
}

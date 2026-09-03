package main

import (
	"log"
	"log/slog"

	"github.com/joho/godotenv"
)

func main() {
	// initialize the application
	Init()

	// load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading .env file", err)
	}

	slog.Info("environment variables loaded successfully")
}

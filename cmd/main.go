package main

import (
	"log"

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
}

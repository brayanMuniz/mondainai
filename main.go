package main

import (
	"log"
	"os"

	"github.com/brayanMuniz/mondainai/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	if geminiApiKey == "" {
		log.Fatal("Error loading gemini api key")
	}

	s, err := server.NewServer(geminiApiKey)
	if err != nil {
		log.Fatal("悲し！")
	}

	log.Fatal(s.Start(":1323"))
}

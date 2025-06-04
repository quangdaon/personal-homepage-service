package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	DSN string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	return &Config{
		DSN: os.Getenv("DATABASE_DSN"),
	}
}

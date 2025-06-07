package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type UpsApiConfig struct {
	BaseUri      string
	ClientId     string
	ClientSecret string
}

type Config struct {
	DSN           string
	LogsDirectory string
	UPSApi        *UpsApiConfig
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	return &Config{
		DSN:           os.Getenv("DATABASE_DSN"),
		LogsDirectory: os.Getenv("LOGS_DIRECTORY"),
		UPSApi: &UpsApiConfig{
			BaseUri:      os.Getenv("UPS_API_BASE_URI"),
			ClientId:     os.Getenv("UPS_API_CLIENT_ID"),
			ClientSecret: os.Getenv("UPS_API_CLIENT_SECRET"),
		},
	}
}

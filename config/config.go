package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string
	DbPath  string
	MainUrl string
}

func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("failed to load .env file: %v", err)

		return nil, err
	}

	cfg := &Config{
		AppEnv:  os.Getenv("APP_ENV"),
		AppPort: os.Getenv("APP_PORT"),
		DbPath:  os.Getenv("DB_PATH"),
		MainUrl: os.Getenv("MAIN_URL"),
	}

	return cfg, nil
}

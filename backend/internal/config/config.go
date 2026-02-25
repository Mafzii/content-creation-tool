package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabasePath string
	Port         string
}

func Load() Config {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "data/app.db"
	}

	return Config{
		DatabasePath: dbPath,
		Port:         port,
	}
}

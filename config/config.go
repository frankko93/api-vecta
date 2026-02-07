package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	API
	Cors
	Database
}

func New() *Config {
	// Intentar cargar .env, si no existe usar defaults
	_ = godotenv.Load()

	return &Config{
		API:      NewAPI(),
		Cors:     NewCors(),
		Database: DataStore(),
	}
}

// GetEnv helper para obtener variable con default
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

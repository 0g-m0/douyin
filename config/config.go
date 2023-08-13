package config

import "os"

// AppConfig stores the application configuration
type AppConfig struct {
	DatabaseName string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	JWTSecretKey string
	// Add more configuration fields here
}

// AppConfigInstance holds the instance of the application configuration
var AppConfigInstance AppConfig

func LoadConfig() {
	AppConfigInstance = AppConfig{
		DatabaseName: getEnv("DATABASE_NAME", "douyin"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "3306"),
		DBUser:       getEnv("DB_USER", "root"),
		DBPassword:   getEnv("DB_PASSWORD", "root"),
		JWTSecretKey: getEnv("JWT_SECREST_KEY", "douyin"),
	}
}

// getEnv retrieves the value of an environment variable or returns a default value if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

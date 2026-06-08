package utils

import (
	"os"
	"strconv"
)

const AppEnvVariable = "GO_ENV"

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvBool retrieves a boolean environment variable with a default fallback.
// If the environment variable is not set or cannot be parsed as a bool, returns defaultValue.
// Accepts: "1", "t", "T", "true", "TRUE", "True", "0", "f", "F", "false", "FALSE", "False"
func GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func GetAppEnv() (env string) {
	return MustGetEnv(AppEnvVariable)
}

func MustGetEnv(key string) string {
	if value := GetEnv(key, ""); value != "" {
		return value
	}
	panic("required environment variable '" + key + "' is not set")
}

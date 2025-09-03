package utils

import "os"

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func MustGetEnv(key string) string {
	if value := GetEnv(key, ""); value != "" {
		return value
	}
	panic("required environment variable '" + key + "' is not set")
}

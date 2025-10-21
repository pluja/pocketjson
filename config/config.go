package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	MasterAPIKey      string
	DefaultMaxSize    int
	AuthenticatedSize int
	DefaultExpiry     time.Duration
	RequestLimit      int
	CORSOrigins       string
	Port              string
	DataDir           string
}

func Load() *Config {
	return &Config{
		MasterAPIKey:      getEnvStr("MASTER_API_KEY", ""),
		DefaultMaxSize:    getEnvInt("DEFAULT_MAX_SIZE", 100*1024),
		AuthenticatedSize: getEnvInt("AUTHENTICATED_MAX_SIZE", 1024*1024),
		DefaultExpiry:     time.Duration(getEnvInt("DEFAULT_EXPIRY_HOURS", 48)) * time.Hour,
		RequestLimit:      getEnvInt("REQUEST_LIMIT", 15),
		CORSOrigins:       getEnvStr("CORS_ALLOWED_ORIGINS", "*"),
		Port:              getEnvStr("PORT", "9819"),
		DataDir:           getEnvStr("DATA_DIR", "data"),
	}
}

func getEnvStr(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

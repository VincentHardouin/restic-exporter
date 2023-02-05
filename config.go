package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type ResticConfig struct {
	Repository string
	Password   string
}

type FeatureToggles struct {
	BackupSummary bool
}

type Config struct {
	Restic         ResticConfig
	FeatureToggles FeatureToggles
}

func getConfig() *Config {
	return &Config{
		Restic: ResticConfig{
			Repository: getEnv("RESTIC_REPOSITORY", "", true),
			Password:   getEnv("RESTIC_PASSWORD", "", true),
		},
		FeatureToggles: FeatureToggles{
			BackupSummary: getEnvAsBool("FT_BACKUP_SUMMARY", false),
		},
	}
}

func getEnv(key string, defaultVal string, required bool) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	if required {
		log.Fatal(fmt.Sprintf("%s is required", key))
	}

	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "", false)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

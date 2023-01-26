package main

import (
	"fmt"
	"log"
	"os"
)

type ResticConfig struct {
	Repository string
	Password   string
}

type Config struct {
	Restic ResticConfig
}

func getConfig() *Config {
	return &Config{
		Restic: ResticConfig{
			Repository: getEnv("RESTIC_REPOSITORY", "", true),
			Password:   getEnv("RESTIC_PASSWORD", "", true),
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

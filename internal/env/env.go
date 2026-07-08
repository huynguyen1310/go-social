package env

import (
	"log"
	"os"
	"strconv"
	"time"
)

func GetString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("warning: invalid %q=%q, using default %d", key, v, fallback)
		return fallback
	}
	return n
}

func GetDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("warning: invalid %q=%q, using default %s", key, v, fallback)
		return fallback
	}
	return d
}

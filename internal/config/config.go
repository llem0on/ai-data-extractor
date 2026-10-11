package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr       string
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string
	LLMTimeout time.Duration
	SchemaPath string
	PDFPages   int
	MaxChars   int
}

func Load() Config {
	return Config{
		Addr:       env("ADDR", ":8090"),
		LLMBaseURL: strings.TrimRight(env("LLM_BASE_URL", ""), "/"),
		LLMAPIKey:  env("LLM_API_KEY", ""),
		LLMModel:   env("LLM_MODEL", ""),
		LLMTimeout: envDuration("LLM_TIMEOUT", 60*time.Second),
		SchemaPath: env("SCHEMA_PATH", ""),
		PDFPages:   envInt("PDF_PAGES", 2),
		MaxChars:   envInt("MAX_CHARS", 6000),
	}
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v, err := strconv.Atoi(env(key, "")); err == nil {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v, err := time.ParseDuration(env(key, "")); err == nil {
		return v
	}
	return def
}

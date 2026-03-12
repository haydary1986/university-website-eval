package config

import "os"

type Config struct {
	DBPath        string
	JWTSecret     string
	ServerPort    string
	UploadDir     string
	DeepSeekKey   string
	GeminiKey     string
	DeepSeekURL   string
	GeminiURL     string
}

func Load() *Config {
	return &Config{
		DBPath:      getEnv("DB_PATH", "./website_eval.db"),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		DeepSeekKey: getEnv("DEEPSEEK_API_KEY", ""),
		GeminiKey:   getEnv("GEMINI_API_KEY", ""),
		DeepSeekURL: getEnv("DEEPSEEK_URL", "https://api.deepseek.com/v1/chat/completions"),
		GeminiURL:   getEnv("GEMINI_URL", "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

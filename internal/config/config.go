package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv, AppPort, AppKey                                 string
	AdminUsername, AdminPassword                            string
	DBHost, DBPort, DBDatabase, DBUsername, DBPassword, SSL string
	SearchEndpoint, OpenAIEndpoint, OpenAIKey, OpenAIModel  string
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load membaca .env (jika ada) lalu variabel lingkungan dengan nilai default.
func Load() Config {
	_ = godotenv.Load()
	return Config{
		AppEnv: getenv("APP_ENV", "development"),
		// APP_PORT diprioritaskan; jika kosong pakai PORT (Render/Railway), lalu 8081.
		AppPort:        getenv("APP_PORT", getenv("PORT", "8081")),
		AppKey:         getenv("APP_KEY", "delcom_blog_secret_key"),
		AdminUsername:  getenv("ADMIN_USERNAME", "admin"),
		AdminPassword:  getenv("ADMIN_PASSWORD", "admin"),
		DBHost:         getenv("DB_HOST", "127.0.0.1"),
		DBPort:         getenv("DB_PORT", "5432"),
		DBDatabase:     getenv("DB_DATABASE", "pabwe_p9_blog"),
		DBUsername:     getenv("DB_USERNAME", "postgres"),
		DBPassword:     getenv("DB_PASSWORD", "postgres"),
		SSL:            getenv("DB_SSLMODE", "disable"),
		SearchEndpoint: getenv("SEARCH_ENGINE_ENDPOINT", "https://delcom.org/api/v1/search"),
		OpenAIEndpoint: getenv("OPENAI_ENDPOINT", "https://delcom.org/api/v1"),
		OpenAIKey:      getenv("OPENAI_KEY", "key"),
		OpenAIModel:    getenv("OPENAI_MODEL", "nama-model"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		c.DBHost, c.DBUsername, c.DBPassword, c.DBDatabase, c.DBPort, c.SSL)
}
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type TelegramConfig struct {
	BotToken      string
	TargetGroupID int64
}

type AppConfig struct {
	AppName string
	Env     string
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Telegram TelegramConfig
}

var cfg *Config

func loadEnv() {
	err := godotenv.Overload()
	if err != nil {
		fmt.Println("⚠️  .env not found, using system env")
	}
}

func mustGet(key string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Printf("❌ Missing required env: %s\n", key)
		os.Exit(1)
	}
	return val
}

func getInt(key string) int {
	val := mustGet(key)
	i, err := strconv.Atoi(val)
	if err != nil {
		fmt.Printf("❌ Invalid int for %s\n", key)
		os.Exit(1)
	}
	return i
}

func getInt64(key string) int64 {
	val := mustGet(key)
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		fmt.Printf("❌ Invalid int64 for %s\n", key)
		os.Exit(1)
	}
	return i
}

func getDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func Load() *Config {
	if cfg != nil {
		return cfg
	}

	loadEnv()

	// --- App ---
	app := AppConfig{
		AppName: getDefault("APP_NAME", "otp-bot"),
		Env:     getDefault("APP_ENV", "development"),
	}

	// --- Database ---
	db := DatabaseConfig{
		Host:     mustGet("DB_HOST"),
		Port:     getInt("DB_PORT"),
		User:     mustGet("DB_USER"),
		Password: mustGet("DB_PASSWORD"),
		Name:     mustGet("DB_NAME"),
		SSLMode:  mustGet("DB_SSLMODE"),
	}

	// --- Telegram ---
	tg := TelegramConfig{
		BotToken:      mustGet("BOT_TOKEN"),
		TargetGroupID: getInt64("TARGET_GROUP_ID"),
	}

	cfg = &Config{
		App:      app,
		Database: db,
		Telegram: tg,
	}

	return cfg
}

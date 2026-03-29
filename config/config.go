package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
	OwnerIDs      []string
	CooldownSecs  int
	EnableWebhook bool
	WebhookURL    string
	ListenPort    int
	VerifyGroup1  string // First group to verify membership
	VerifyGroup2  string // Second group to verify membership
	VerifyURL1    string // First group join URL
	VerifyURL2    string // Second group join URL
}

type AppConfig struct {
	AppName string
	Env     string
}

type ScraperConfig struct {
	LoginURL string
	SMSURL   string
	Username string
	Password string
}

type RedisConfig struct {
	Addr       string
	Password   string
	DB         int
	KeyPrefix  string
	ActiveTTL  int
	EnableTLS  bool
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Telegram TelegramConfig
	Scraper  ScraperConfig
	Redis    RedisConfig
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

func getDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getBoolDefault(key string, def bool) bool {
	val := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if val == "" {
		return def
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return def
	}
	return parsed
}

// parseInt64List parses a comma-separated list of int64 values
func parseInt64List(raw string) []int64 {
	var result []int64
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

// parseStringList parses a comma-separated list of strings
func parseStringList(raw string) []string {
	var result []string
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
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
	dbCfg := DatabaseConfig{
		Host:     mustGet("DB_HOST"),
		Port:     getInt("DB_PORT"),
		User:     mustGet("DB_USER"),
		Password: mustGet("DB_PASSWORD"),
		Name:     mustGet("DB_NAME"),
		SSLMode:  mustGet("DB_SSLMODE"),
	}

	// --- Telegram ---
	ownerIDsRaw := mustGet("OWNER_IDS")
	cooldown, _ := strconv.Atoi(getDefault("COOLDOWN_SECONDS", "10"))
	listenPort, _ := strconv.Atoi(getDefault("LISTEN_PORT", "8080"))

	tg := TelegramConfig{
		BotToken:      mustGet("BOT_TOKEN"),
		OwnerIDs:      parseStringList(ownerIDsRaw),
		CooldownSecs:  cooldown,
		EnableWebhook: getBoolDefault("ENABLE_WEBHOOK", false),
		WebhookURL:    os.Getenv("WEBHOOK_URL"),
		ListenPort:    listenPort,
		VerifyGroup1:  getDefault("VERIFY_GROUP_1", ""),
		VerifyGroup2:  getDefault("VERIFY_GROUP_2", ""),
		VerifyURL1:    getDefault("VERIFY_URL_1", ""),
		VerifyURL2:    getDefault("VERIFY_URL_2", ""),
	}

	// --- Scraper ---
	scraper := ScraperConfig{
		LoginURL: getDefault("SCRAPER_LOGIN_URL", "http://185.2.83.39/ints/login"),
		SMSURL:   getDefault("SCRAPER_SMS_URL", "http://185.2.83.39/ints/agent/SMSCDRReports"),
		Username: os.Getenv("SCRAPER_USERNAME"),
		Password: os.Getenv("SCRAPER_PASSWORD"),
	}

	// --- Redis ---
	redis := RedisConfig{
		Addr:      getDefault("REDIS_ADDR", "127.0.0.1:6379"),
		Password:  getDefault("REDIS_PASSWORD", ""),
		DB:        func() int { v, _ := strconv.Atoi(getDefault("REDIS_DB", "0")); return v }(),
		KeyPrefix: getDefault("REDIS_KEY_PREFIX", "sharkbot"),
		ActiveTTL: func() int { v, _ := strconv.Atoi(getDefault("REDIS_ACTIVE_TTL_SECONDS", "7200")); return v }(),
		EnableTLS: getBoolDefault("REDIS_ENABLE_TLS", false),
	}

	cfg = &Config{
		App:      app,
		Database: dbCfg,
		Telegram: tg,
		Scraper:  scraper,
		Redis:    redis,
	}

	return cfg
}

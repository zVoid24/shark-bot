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
	VerifyGroup3  string // Third group to verify membership
	VerifyURL1    string // First group join URL
	VerifyURL2    string // Second group join URL
	VerifyURL3    string // Third group join URL
	OTPTargetChatID int64  // Group ID where OTPs will be forwarded
	ListenEnabled   bool   // Whether to poll for updates
}

type AppConfig struct {
	AppName string
	Env     string
}

type ScraperAccount struct {
	Username string
	Password string
	LoginURL string // Optional: overrides global LoginURL
	SMSURL   string // Optional: overrides global SMSURL
}

type ScraperConfig struct {
	LoginURL string
	SMSURL   string
	Accounts []ScraperAccount
}

type RedisConfig struct {
	Addr      string
	Password  string
	DB        int
	KeyPrefix string
	ActiveTTL int
	EnableTLS bool
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

func parseScraperAccounts(raw string) []ScraperAccount {
	var accounts []ScraperAccount
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		// Format: username:password@loginURL|smsURL
		// Or: username:password
		var loginURL, smsURL string
		if atIdx := strings.Index(s, "@"); atIdx != -1 {
			urls := s[atIdx+1:]
			s = s[:atIdx]
			urlParts := strings.SplitN(urls, "|", 2)
			if len(urlParts) >= 1 {
				loginURL = urlParts[0]
			}
			if len(urlParts) >= 2 {
				smsURL = urlParts[1]
			}
		}

		parts := strings.SplitN(s, ":", 2)
		if len(parts) == 2 {
			accounts = append(accounts, ScraperAccount{
				Username: parts[0],
				Password: parts[1],
				LoginURL: loginURL,
				SMSURL:   smsURL,
			})
		}
	}
	return accounts
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
		VerifyGroup3:  getDefault("VERIFY_GROUP_3", ""),
		VerifyURL1:    getDefault("VERIFY_URL_1", ""),
		VerifyURL2:    getDefault("VERIFY_URL_2", ""),
		VerifyURL3:    getDefault("VERIFY_URL_3", ""),
		OTPTargetChatID: func() int64 { v, _ := strconv.ParseInt(os.Getenv("OTP_TARGET_CHAT_ID"), 10, 64); return v }(),
		ListenEnabled:   getBoolDefault("TELEGRAM_LISTEN_ENABLED", true),
	}

	// --- Scraper ---
	accounts := parseScraperAccounts(os.Getenv("SCRAPER_ACCOUNTS"))
	// Fallback to legacy single account env vars if list is empty
	if len(accounts) == 0 {
		user := os.Getenv("SCRAPER_USERNAME")
		pass := os.Getenv("SCRAPER_PASSWORD")
		if user != "" && pass != "" {
			accounts = append(accounts, ScraperAccount{Username: user, Password: pass})
		}
	}

	scraper := ScraperConfig{
		LoginURL: getDefault("SCRAPER_LOGIN_URL", "http://185.2.83.39/ints/login"),
		SMSURL:   getDefault("SCRAPER_SMS_URL", "http://185.2.83.39/ints/agent/SMSCDRReports"),
		Accounts: accounts,
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

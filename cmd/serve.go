package cmd

import (
	"context"
	"crypto/tls"
	"shark_bot/infra/db"
	"shark_bot/infra/repository"
	"shark_bot/internal/activenumber"
	"shark_bot/internal/admin"
	"shark_bot/internal/number"
	"shark_bot/internal/processednumber"
	"shark_bot/internal/seennumber"
	"shark_bot/internal/settings"
	"shark_bot/internal/stats"
	"shark_bot/internal/user"
	"shark_bot/pkg/logger"

	"shark_bot/bot"
	"shark_bot/config"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

func Serve() {
	log := logger.New("serve")

	// 1. Load config
	cnf := config.Load()

	// 2. Connect to PostgreSQL
	dbConn, err := db.NewConnection(&cnf.Database)
	if err != nil {
		log.Error("unable to connect to database", "err", err)
		panic(err)
	}
	defer dbConn.Close()

	// 3. Run migrations
	if err := db.Migrate(dbConn); err != nil {
		log.Error("migration failed", "err", err)
		panic(err)
	}

	// 4. Build concrete repos (infra layer)
	userRepo := repository.NewUserRepo(dbConn)
	adminRepo := repository.NewAdminRepo(dbConn)
	numberRepo := repository.NewNumberRepo(dbConn)
	activeRepo := repository.NewActiveNumberRepo(dbConn)
	settingsRepo := repository.NewSettingsRepo(dbConn)
	statsRepo := repository.NewStatsRepo(dbConn)
	seenRepo := repository.NewSeenNumberRepo(dbConn)
	processedRepo := repository.NewProcessedNumberRepo(dbConn)

	// 5. Wrap repos in domain services (application layer)
	userSvc := user.NewService(userRepo)
	adminSvc := admin.NewService(adminRepo)
	numberSvc := number.NewService(numberRepo)
	activeSvc := activenumber.NewService(activeRepo)
	settingsSvc := settings.NewService(settingsRepo)
	statsSvc := stats.NewService(statsRepo)
	seenSvc := seennumber.NewService(seenRepo)
	processedSvc := processednumber.NewService(processedRepo)

	// 5.2 Initialize Redis and active-number cache (optional fallback to DB if unavailable)
	var redisClient *redis.Client
	var activeCache *bot.ActiveNumberCache
	redisOpts := &redis.Options{
		Addr:     cnf.Redis.Addr,
		Password: cnf.Redis.Password,
		DB:       cnf.Redis.DB,
	}
	if cnf.Redis.EnableTLS {
		redisOpts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	redisClient = redis.NewClient(redisOpts)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Warn("redis unavailable; continuing with DB-only fallback", "addr", cnf.Redis.Addr, "err", err)
		redisClient = nil
	} else {
		activeCache = bot.NewActiveNumberCache(
			redisClient,
			cnf.Redis.KeyPrefix,
			time.Duration(cnf.Redis.ActiveTTL)*time.Second,
		)
		log.Info("redis connected", "addr", cnf.Redis.Addr)
	}

	// 5.3 Initialize Verification Cache (uses Redis if available)
	var verCache *bot.VerificationCache
	if redisClient != nil {
		verCache = bot.NewVerificationCache(
			redisClient,
			cnf.Redis.KeyPrefix,
			2*time.Hour, // 2 hour TTL for verification status
		)
	}

	// 5.5 Initialize Scrapers
	var scrapers []*bot.Scraper
	for _, acc := range cnf.Scraper.Accounts {
		scrp, err := bot.NewScraper(cnf.Scraper.LoginURL, cnf.Scraper.SMSURL, acc.Username, acc.Password)
		if err != nil {
			log.Error("failed to init scraper", "user", acc.Username, "err", err)
			continue
		}
		scrapers = append(scrapers, scrp)
	}

	// 6. Seed initial owner IDs as admins
	if err := adminSvc.SeedOwners(cnf.Telegram.OwnerIDs); err != nil {
		log.Warn("could not seed owners", "err", err)
	}

	// 7. Connect to Telegram
	api, err := tgbotapi.NewBotAPI(cnf.Telegram.BotToken)
	if err != nil {
		log.Error("telegram bot init failed", "err", err)
		panic(err)
	}

	// 8. Build and start bot
	b := bot.New(
		api,
		userSvc,
		adminSvc,
		numberSvc,
		activeSvc,
		settingsSvc,
		statsSvc,
		seenSvc,
		processedSvc,
		scrapers,
		redisClient,
		activeCache,
		verCache,
		cnf.Telegram.OwnerIDs,
		cnf.Telegram.CooldownSecs,
		cnf.Telegram.VerifyGroup1,
		cnf.Telegram.VerifyGroup2,
		cnf.Telegram.VerifyGroup3,
		cnf.Telegram.VerifyURL1,
		cnf.Telegram.VerifyURL2,
		cnf.Telegram.VerifyURL3,
	)

	if cnf.Telegram.EnableWebhook && cnf.Telegram.WebhookURL != "" {
		b.StartWebhook(cnf.Telegram.WebhookURL, cnf.Telegram.ListenPort)
	} else {
		b.Start()
	}
}

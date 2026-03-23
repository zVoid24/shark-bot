package cmd

import (
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

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	// 5.5 Initialize Scraper
	scrp, err := bot.NewScraper(cnf.Scraper.LoginURL, cnf.Scraper.SMSURL, cnf.Scraper.Username, cnf.Scraper.Password)
	if err != nil {
		log.Error("failed to init scraper", "err", err)
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
		scrp,
		cnf.Telegram.TargetGroupIDs,
		cnf.Telegram.OwnerIDs,
		cnf.Telegram.CooldownSecs,
	)

	b.Start()
}

package bot

//testing ci/cd pipeline
import (
	"context"
	"fmt"
	"net/http"
	"time"
	"shark_bot/internal/activenumber"
	"shark_bot/internal/admin"
	"shark_bot/internal/number"
	"shark_bot/internal/processednumber"
	"shark_bot/internal/seennumber"
	"shark_bot/internal/settings"
	"shark_bot/internal/stats"
	"shark_bot/internal/user"
	"shark_bot/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

var log = logger.New("bot")

// Temporary switch: keep scraper code intact but do not run the background worker.
const otpWorkerEnabled = true

// Bot aggregates all dependencies via domain service interfaces.
type Bot struct {
	api          *tgbotapi.BotAPI
	userSvc      *user.Service
	adminSvc     *admin.Service
	numberSvc    *number.Service
	activeSvc    *activenumber.Service
	settingsSvc  *settings.Service
	statsSvc     *stats.Service
	seenSvc      *seennumber.Service
	processedSvc *processednumber.Service
	scrapers     []*Scraper
	crapiClient  *CRAPIClient
	redisClient  *redis.Client
	activeCache  *ActiveNumberCache
	verifyCache  *VerificationCache
	ownerIDs     []string
	cooldownSecs int
	verifyGroup1 string // First group URL/ID to verify membership
	verifyGroup2 string // Second group URL/ID to verify membership
	verifyGroup3 string // Third group URL/ID to verify membership
	verifyURL1   string // First group join redirection URL
	verifyURL2   string // Second group join redirection URL
	verifyURL3   string // Third group join redirection URL
	// Conversation state per user (for add/remove number flow)
	convState map[int64]*convContext
}

// otpMessage is removed as it's no longer needed for group chats.

// New creates the Bot instance with service-layer dependencies.
func New(
	api *tgbotapi.BotAPI,
	userSvc *user.Service,
	adminSvc *admin.Service,
	numberSvc *number.Service,
	activeSvc *activenumber.Service,
	settingsSvc *settings.Service,
	statsSvc *stats.Service,
	seenSvc *seennumber.Service,
	processedSvc *processednumber.Service,
	scrapers []*Scraper,
	crapiClient *CRAPIClient,
	redisClient *redis.Client,
	activeCache *ActiveNumberCache,
	verifyCache *VerificationCache,
	ownerIDs []string,
	cooldownSecs int,
	verifyGroup1 string,
	verifyGroup2 string,
	verifyGroup3 string,
	verifyURL1 string,
	verifyURL2 string,
	verifyURL3 string,
) *Bot {
	return &Bot{
		api:          api,
		userSvc:      userSvc,
		adminSvc:     adminSvc,
		numberSvc:    numberSvc,
		activeSvc:    activeSvc,
		settingsSvc:  settingsSvc,
		statsSvc:     statsSvc,
		seenSvc:      seenSvc,
		processedSvc: processedSvc,
		scrapers:     scrapers,
		crapiClient:  crapiClient,
		redisClient:  redisClient,
		activeCache:  activeCache,
		verifyCache:  verifyCache,
		ownerIDs:     ownerIDs,
		cooldownSecs: cooldownSecs,
		verifyGroup1:  verifyGroup1,
		verifyGroup2:  verifyGroup2,
		verifyGroup3:  verifyGroup3,
		verifyURL1:    verifyURL1,
		verifyURL2:    verifyURL2,
		verifyURL3:    verifyURL3,
		convState:    make(map[int64]*convContext),
	}
}

// Start begins the polling loop.
func (b *Bot) Start() {
	b.api.Debug = false
	if _, err := b.api.Request(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true}); err != nil {
		log.Warn("failed to delete existing webhook before polling", "err", err)
	} else {
		log.Info("cleared webhook before polling")
	}
	log.Info("bot started in polling mode", "username", b.api.Self.UserName)

	if otpWorkerEnabled {
		b.seedActiveCacheFromDB()
		go b.otpWorker()
	} else {
		log.Info("OTP worker disabled")
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.handleUpdate(update)
	}
}

// StartWebhook starts the bot using Telegram webhooks.
func (b *Bot) StartWebhook(webhookURL string, port int) {
	b.api.Debug = false
	log.Info("bot starting in webhook mode", "username", b.api.Self.UserName, "url", webhookURL, "port", port)

	if otpWorkerEnabled {
		b.seedActiveCacheFromDB()
		go b.otpWorker()
	} else {
		log.Info("OTP worker disabled")
	}

	wh, _ := tgbotapi.NewWebhook(webhookURL)
	_, err := b.api.Request(wh)
	if err != nil {
		log.Error("failed to set webhook", "err", err)
		panic(err)
	}

	updates := b.api.ListenForWebhook("/")
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	for update := range updates {
		log.Info("<- incoming update", "id", update.UpdateID)
		go b.handleUpdate(update)
	}
}

func (b *Bot) seedActiveCacheFromDB() {
	if b.activeCache == nil {
		return
	}
	all, err := b.activeSvc.GetAll()
	if err != nil {
		log.Warn("failed to load active numbers for redis seed", "err", err)
		return
	}
	ctx := context.Background()
	for _, an := range all {
		if err := b.activeCache.Set(ctx, an); err != nil {
			log.Warn("failed to seed active number in redis", "number", an.Number, "user_id", an.UserID, "err", err)
		}
	}
}

// handleUpdate routes each Telegram update and logs it.
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	switch {
	case update.CallbackQuery != nil:
		log.Info("callback", "user", update.CallbackQuery.From.ID, "data", update.CallbackQuery.Data)
		b.handleCallback(update.CallbackQuery)

	case update.Message != nil && update.Message.Chat.IsPrivate():
		start := time.Now()
		log.Info("<- incoming private message", "user", update.Message.From.ID, "text", update.Message.Text)
		
		b.handlePrivateMessage(update.Message)
		
		log.Info("-> processed private message", "user", update.Message.From.ID, "duration", time.Since(start).String())
	}
}

// handlePrivateMessage routes private messages to the correct handler.
func (b *Bot) handlePrivateMessage(msg *tgbotapi.Message) {
	b.trackKnownUser(msg.From)
	userID := fmt.Sprintf("%d", msg.From.ID)

	if msg.Text != "" && msg.Document == nil {
		if b.handleConversationText(msg) {
			return
		}
	}
	if msg.Document != nil {
		if b.handleConversationDocument(msg) {
			return
		}
	}

	if msg.IsCommand() {
		cmd := msg.Command()
		if cmd == "start" {
			b.handleStart(msg)
			return
		}

		// Check if admin or verified
		isAdmin, _ := b.adminSvc.IsAdmin(userID)
		if !isAdmin && !b.isUserVerified(msg.From.ID) {
			b.showVerificationScreen(msg.Chat.ID)
			return
		}

		b.handleCommand(msg)
		return
	}

	// Check if user is verified for non-command messages
	isAdmin, _ := b.adminSvc.IsAdmin(userID)

	if !isAdmin && !b.isUserVerified(msg.From.ID) {
		// Non-admin users must be verified
		b.showVerificationScreen(msg.Chat.ID)
		return
	}

	switch msg.Text {
	case "Get a Phone Number ☎️":
		b.handleGetNumber(msg)
	case "📊 My Status":
		b.handleMyStatus(msg)
	}
}

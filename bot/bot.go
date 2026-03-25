package bot

//testing ci/cd pipeline
import (
	"fmt"
	"net/http"
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
)

var log = logger.New("bot")

// Temporary switch: keep scraper code intact but do not run the background worker.
const otpWorkerEnabled = false

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
	scraper      *Scraper
	ownerIDs     []string
	cooldownSecs int
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
	scraper *Scraper,
	ownerIDs []string,
	cooldownSecs int,
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
		scraper:      scraper,
		ownerIDs:     ownerIDs,
		cooldownSecs: cooldownSecs,
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

// handleUpdate routes each Telegram update and logs it.
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	switch {
	case update.CallbackQuery != nil:
		log.Info("callback", "user", update.CallbackQuery.From.ID, "data", update.CallbackQuery.Data)
		b.handleCallback(update.CallbackQuery)

	case update.Message != nil && update.Message.Chat.IsPrivate():
		log.Info("private message", "user", update.Message.From.ID, "text", update.Message.Text)
		b.handlePrivateMessage(update.Message)
	}
}

// handlePrivateMessage routes private messages to the correct handler.
func (b *Bot) handlePrivateMessage(msg *tgbotapi.Message) {
	b.trackKnownUser(msg.From)

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
		b.handleCommand(msg)
		return
	}

	switch msg.Text {
	case "Get a Phone Number ☎️":
		b.handleGetNumber(msg)
	case "📊 My Status":
		b.handleMyStatus(msg)
	}
}

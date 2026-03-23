package bot

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

// workerPoolSize caps the number of Telegram update handlers running at the
// same time.  This prevents goroutine/memory exhaustion on low-spec VPS hosts
// while still handling bursts smoothly.
const workerPoolSize = 50

var log = logger.New("bot")

// Bot aggregates all dependencies via domain service interfaces.
type Bot struct {
	api            *tgbotapi.BotAPI
	userSvc        *user.Service
	adminSvc       *admin.Service
	numberSvc      *number.Service
	activeSvc      *activenumber.Service
	settingsSvc    *settings.Service
	statsSvc       *stats.Service
	seenSvc        *seennumber.Service
	processedSvc   *processednumber.Service
	scraper        *Scraper
	targetGroupIDs map[int64]bool
	ownerIDs       []string
	cooldownSecs   int
	otpChan        chan otpMessage
	// Conversation state per user (for add/remove number flow)
	convState map[int64]*convContext
}

type otpMessage struct {
	text        string
	chatID      int64
	messageID   int
	replyMarkup *tgbotapi.InlineKeyboardMarkup
}

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
	targetGroupIDs []int64,
	ownerIDs []string,
	cooldownSecs int,
) *Bot {
	groupMap := make(map[int64]bool)
	for _, id := range targetGroupIDs {
		groupMap[id] = true
	}
	return &Bot{
		api:            api,
		userSvc:        userSvc,
		adminSvc:       adminSvc,
		numberSvc:      numberSvc,
		activeSvc:      activeSvc,
		settingsSvc:    settingsSvc,
		statsSvc:       statsSvc,
		seenSvc:        seenSvc,
		processedSvc:   processedSvc,
		scraper:        scraper,
		targetGroupIDs: groupMap,
		ownerIDs:       ownerIDs,
		cooldownSecs:   cooldownSecs,
		otpChan:        make(chan otpMessage, 256),
		convState:      make(map[int64]*convContext),
	}
}

// Start begins the polling loop and background workers.
// It blocks until a SIGINT or SIGTERM signal is received, then performs a
// graceful shutdown: it stops accepting new updates and waits for all
// in-flight handlers to finish before returning.
func (b *Bot) Start() {
	b.api.Debug = false
	log.Info("bot started", "username", b.api.Self.UserName)

	// ctx is cancelled when a shutdown signal arrives, which stops all
	// background goroutines (OTP worker excluded — it also exits on process
	// termination, but the conv-cleaner uses the context explicitly).
	ctx, cancel := context.WithCancel(context.Background())

	go b.otpWorker()
	go b.startConvCleaner(ctx)

	// Stop the Telegram polling loop when the OS asks us to shut down.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Info("shutdown signal received", "signal", sig.String())
		cancel()                         // stop background goroutines
		b.api.StopReceivingUpdates()     // close the updates channel
	}()

	// Semaphore that limits concurrent update handlers.
	sem := make(chan struct{}, workerPoolSize)
	var wg sync.WaitGroup

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		sem <- struct{}{}
		wg.Add(1)
		go func(upd tgbotapi.Update) {
			// Defers run LIFO: semaphore slot is freed first, then the
			// WaitGroup is decremented, so capacity is restored before
			// signalling completion to any wg.Wait() caller.
			defer wg.Done()
			defer func() { <-sem }()
			b.handleUpdate(upd)
		}(update)
	}

	// Wait for all in-flight handlers to complete before the process exits.
	wg.Wait()
	log.Info("all update handlers finished, bot stopped cleanly")
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

	case update.Message != nil:
		log.Info("group message", "chat", update.Message.Chat.ID, "user", update.Message.From.ID)
		b.handleGroupMessage(update.Message)

	case update.ChannelPost != nil:
		log.Info("channel post", "chat", update.ChannelPost.Chat.ID)
		b.handleGroupMessage(update.ChannelPost)

	case update.EditedMessage != nil && !update.EditedMessage.Chat.IsPrivate():
		log.Info("edited group message", "chat", update.EditedMessage.Chat.ID)
		b.handleGroupMessage(update.EditedMessage)

	case update.EditedChannelPost != nil:
		log.Info("edited channel post", "chat", update.EditedChannelPost.Chat.ID)
		b.handleGroupMessage(update.EditedChannelPost)
	}
}

// handlePrivateMessage routes private messages to the correct handler.
func (b *Bot) handlePrivateMessage(msg *tgbotapi.Message) {
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

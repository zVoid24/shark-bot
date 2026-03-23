package bot

import (
	"shark_bot/internal/activenumber"
	"shark_bot/internal/admin"
	"shark_bot/internal/number"
	"shark_bot/internal/seennumber"
	"shark_bot/internal/settings"
	"shark_bot/internal/stats"
	"shark_bot/internal/user"
	"shark_bot/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
		targetGroupIDs: groupMap,
		ownerIDs:       ownerIDs,
		cooldownSecs:   cooldownSecs,
		otpChan:        make(chan otpMessage, 256),
		convState:      make(map[int64]*convContext),
	}
}

// Start begins the polling loop and background workers.
func (b *Bot) Start() {
	b.api.Debug = false
	log.Info("bot started", "username", b.api.Self.UserName)

	go b.otpWorker()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.handleUpdate(update)
	}
}

// handleUpdate routes each Telegram update and logs it.
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	log.Info("incoming update detected", "id", update.UpdateID, "has_message", update.Message != nil)
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

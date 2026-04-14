package bot

import (
	"fmt"
	"shark_bot/pkg/logger"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStart registers user and shows verification screen
func (b *Bot) handleStart(msg *tgbotapi.Message) {
	user := msg.From
	if user == nil {
		return
	}
	userID := fmt.Sprintf("%d", user.ID)

	if blocked, _ := b.userSvc.IsBlocked(userID); blocked {
		return
	}

	fullName := user.FirstName
	if user.LastName != "" {
		fullName += " " + user.LastName
	}

	isNew, err := b.userSvc.EnsureUser(userID, fullName)
	if err != nil {
		logger.L.Error("EnsureUser failed", "err", err)
	}

	// Notify owners of new user
	if isNew {
		username := user.UserName
		if username == "" {
			username = "N/A"
		}
		notif := fmt.Sprintf(
			"<b>👤 New User Joined!</b>\n\n<b>Name:</b> %s\n<b>ID:</b> <code>%s</code>\n<b>Username:</b> @%s",
			fullName, userID, username)
		for _, ownerID := range b.ownerIDs {
			var oid int64
			fmt.Sscanf(ownerID, "%d", &oid)
			if oid != 0 {
				go func(id int64) {
					m := tgbotapi.NewMessage(id, notif)
					m.ParseMode = tgbotapi.ModeHTML
					b.api.Send(m)
				}(oid)
			}
		}
	}

	// Check if user is verified (joined both groups)
	if !b.isUserVerified(user.ID) {
		// Show verification screen
		b.showVerificationScreen(msg.Chat.ID)
		return
	}

	// User is verified, show welcome message
	welcome := fmt.Sprintf(
		"<b>👋 Welcome back, %s!</b>\n\n"+
			"🚀 <b>Everything is ready for you to start.</b>\n\n"+
			"📱 Click the button below to get your phone numbers and start receiving OTPs instantly.\n\n"+
			"🎥 <i>Need help? Check out our professional guide on YouTube!</i>",
		fullName)

	markup := CustomMarkup{
		InlineKeyboard: [][]CustomButton{
			{
				{
					Text:          "📺 Watch Guide",
					URL:           "https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz",
					CustomEmojiID: "5942902988564600402",
				},
			},
		},
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, welcome)
	reply.ParseMode = tgbotapi.ModeHTML
	reply.ReplyMarkup = markup
	if _, err := b.api.Send(reply); err != nil {
		logger.L.Error("failed to send welcome message", "err", err)
	}


	// Also send the main reply keyboard to ensure it's visible
	keyboardMsg := tgbotapi.NewMessage(msg.Chat.ID, "⬇️ <b>Use the menu below to navigate:</b>")
	keyboardMsg.ParseMode = tgbotapi.ModeHTML
	keyboardMsg.ReplyMarkup = mainKeyboard()
	b.api.Send(keyboardMsg)
}

// ---- User cooldown tracking ----
var (
	userCooldowns   = make(map[string]int64)
	userCooldownsMu sync.Mutex
)

func (b *Bot) checkCooldown(userID string) (bool, int) {
	userCooldownsMu.Lock()
	defer userCooldownsMu.Unlock()
	last, ok := userCooldowns[userID]
	if !ok {
		return true, 0
	}
	now := time.Now().Unix()
	diff := int(int64(b.cooldownSecs) - (now - last))
	if diff > 0 {
		return false, diff
	}
	return true, 0
}

func (b *Bot) setCooldown(userID string) {
	userCooldownsMu.Lock()
	defer userCooldownsMu.Unlock()
	userCooldowns[userID] = time.Now().Unix()
}

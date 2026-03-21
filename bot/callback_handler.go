package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCallback routes all inline button presses
func (b *Bot) handleCallback(cb *tgbotapi.CallbackQuery) {
	// Always answer to remove the loading spinner
	b.answerCallback(cb.ID, "", false)

	userID := fmt.Sprintf("%d", cb.From.ID)
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID
	data := cb.Data

	// Block check
	if blocked, _ := b.userSvc.IsBlocked(userID); blocked {
		b.answerCallback(cb.ID, "You are blocked.", true)
		return
	}

	switch {
	case data == "back_to_platforms":
		b.showPlatformList(chatID, msgID, true)

	case strings.HasPrefix(data, "select_platform::"):
		platform := strings.TrimPrefix(data, "select_platform::")
		b.showCountryList(chatID, msgID, platform)

	case strings.HasPrefix(data, "back_to_countries::"):
		platform := strings.TrimPrefix(data, "back_to_countries::")
		b.showCountryList(chatID, msgID, platform)

	case strings.HasPrefix(data, "select_country::"):
		parts := strings.Split(data, "::")
		if len(parts) == 3 {
			b.assignNumbers(chatID, cb.From.ID, parts[1], parts[2], msgID, false)
		}

	case strings.HasPrefix(data, "change_number::"):
		// Cooldown check
		ok, remaining := b.checkCooldown(userID)
		if !ok {
			b.answerCallback(cb.ID, fmt.Sprintf("Please wait %d seconds.", remaining), true)
			return
		}
		parts := strings.Split(data, "::")
		if len(parts) == 3 {
			b.setCooldown(userID)
			b.assignNumbers(chatID, cb.From.ID, parts[1], parts[2], msgID, true)
		}
	}
}

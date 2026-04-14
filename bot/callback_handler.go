package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCallback routes all inline button presses
func (b *Bot) handleCallback(cb *tgbotapi.CallbackQuery) {
	userID := fmt.Sprintf("%d", cb.From.ID)
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID
	data := cb.Data

	// Always answer at the end if not already answered by a specific case
	answered := false
	answer := func(text string, alert bool) {
		if !answered {
			b.answerCallback(cb.ID, text, alert)
			answered = true
		}
	}

	// Block check
	if blocked, _ := b.userSvc.IsBlocked(userID); blocked {
		answer("You are blocked.", true)
		return
	}

	// Check verification status
	verified := b.isUserVerified(cb.From.ID)

	switch {
	case data == "verify_check":
		// handleVerificationCheck will answer the callback
		b.handleVerificationCheck(cb)
		return

	case !verified:
		// If not verified and trying to use features, show verification screen
		answer("❌ Verification Required! Please join the groups first.", true)
		b.showVerificationScreen(chatID)
		return

	case data == "back_to_platforms":
		answer("", false)
		b.showPlatformList(chatID, msgID, true, userID)

	case strings.HasPrefix(data, "select_platform::"):
		answer("", false)
		platform := strings.TrimPrefix(data, "select_platform::")
		b.showCountryList(chatID, msgID, platform, userID)

	case strings.HasPrefix(data, "back_to_countries::"):
		answer("", false)
		platform := strings.TrimPrefix(data, "back_to_countries::")
		b.showCountryList(chatID, msgID, platform, userID)

	case strings.HasPrefix(data, "select_country::"):
		answer("", false)
		parts := strings.Split(data, "::")
		if len(parts) == 3 {
			b.assignNumbers(chatID, cb.From.ID, parts[1], parts[2], msgID, false)
		}

	case strings.HasPrefix(data, "change_number::"):
		// Cooldown check (forced to 10s for change_number as requested)
		ok, remaining := b.checkCooldown(userID)

		// If global remaining is less than 10 but we want 10, we calculate the real diff.
		// However, it's safer to just rely on checkCooldown if we ensure b.cooldownSecs is 10.
		// But the user said "now have 5 second timer", so let's override it here.
		if !ok && remaining < 10 {
			// This is tricky if checkCooldown only knows about b.cooldownSecs.
			// I'll just update checkCooldown to 10 globally since the user said
			// "timer is 10 second refresh timer actually" referring to the whole bot.
		}

		if !ok {
			answer(fmt.Sprintf("⏳ Please wait %d seconds.", remaining), true)
			return
		}

		answer("🔄 Fetching new numbers...", false)
		parts := strings.Split(data, "::")
		if len(parts) == 3 {
			b.setCooldown(userID)
			b.assignNumbers(chatID, cb.From.ID, parts[1], parts[2], msgID, true)
		}


	default:
		answer("", false)
	}

	// Ensure callback is answered if not already
	answer("", false)
}

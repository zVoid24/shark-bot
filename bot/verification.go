package bot

import (
	"context"
	"fmt"
	"shark_bot/pkg/logger"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// verificationKeyboard returns the verification inline keyboard with join buttons
func (b *Bot) verificationKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📱 Join Group 1", b.verifyURL1),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("⛧ Join Group 2", b.verifyURL2),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Check Verification", "verify_check"),
		),
	)
}

// showVerificationScreen shows the verification message with buttons
func (b *Bot) showVerificationScreen(chatID int64) {
	msg := `<b>🔐 Verification Required</b>

Before you can use this bot, you must join both of our groups:

<b>📱 Group 1:</b> 𝗦𝗛𝗔𝗥𝗞 𝗦𝗠𝗦 𝗕𝗔𝗖𝗞𝗨𝗣
<b>⛧ Group 2:</b> 𝙎𝙃𝘼𝙍𝙆 𝙈𝙀𝙏𝙃𝙊𝘿 ⛧

Join both groups using the buttons below, then click "Check Verification" to verify your membership.`

	b.sendHTMLWithMarkup(chatID, msg, b.verificationKeyboard())
}

// isUserVerified checks if a user has joined both required groups
func (b *Bot) isUserVerified(userID int64) bool {
	ctx := context.Background()
	// 1. Check cache first
	if b.verifyCache != nil && b.verifyCache.IsVerified(ctx, userID) {
		return true
	}

	// 2. Perform membership checks via Telegram API
	isVerified := b.performMembershipChecks(userID)

	// 3. Cache the result if verified
	if isVerified && b.verifyCache != nil {
		if err := b.verifyCache.SetVerified(ctx, userID); err != nil {
			logger.L.Warn("failed to cache verification status", "user_id", userID, "err", err)
		}
	}

	return isVerified
}

func (b *Bot) performMembershipChecks(userID int64) bool {
	// Check membership in first group
	if !b.checkGroupMembership(userID, b.verifyGroup1) {
		return false
	}

	// Check membership in second group
	if !b.checkGroupMembership(userID, b.verifyGroup2) {
		return false
	}

	return true
}

// checkGroupMembership checks if a user is a member of a group/channel
func (b *Bot) checkGroupMembership(userID int64, groupIdentifier string) bool {
	if groupIdentifier == "" || groupIdentifier == "0" {
		return true
	}

	identifier := groupIdentifier

	// Parse numeric ID or username
	var chatID int64
	var username string

	if strings.HasPrefix(identifier, "@") {
		username = identifier
	} else if strings.HasPrefix(identifier, "-") {
		// Numeric ID
		fmt.Sscanf(identifier, "%d", &chatID)
		// Ensure -100 prefix for supergroups if not present
		if chatID > -1000000000000 && chatID < 0 {
			s := identifier
			if !strings.HasPrefix(s, "-100") {
				s = "-100" + strings.TrimPrefix(s, "-")
				fmt.Sscanf(s, "%d", &chatID)
			}
		}
	} else {
		// Default to username
		username = "@" + identifier
	}

	cfg := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	}
	if chatID == 0 {
		cfg.SuperGroupUsername = username
	}

	member, err := b.api.GetChatMember(cfg)
	if err != nil {
		logger.L.Error("Failed to check group membership",
			"user_id", userID,
			"group", identifier,
			"resolved_id", chatID,
			"resolved_user", username,
			"err", err)
		return false
	}

	status := member.Status
	logger.L.Debug("Membership status checked",
		"user_id", userID,
		"group", identifier,
		"status", status)

	// Valid statuses for being a "member"
	return status == "member" || status == "administrator" || status == "creator" || status == "restricted"
}

// handleVerificationCheck handles the verification check button callback
func (b *Bot) handleVerificationCheck(cb *tgbotapi.CallbackQuery) {
	ctx := context.Background()
	userID := cb.From.ID
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID

	// Clear cache to force a fresh check from Telegram
	if b.verifyCache != nil {
		_ = b.verifyCache.Clear(ctx, userID)
	}

	if b.isUserVerified(userID) {
		// Show success pop-up alert
		b.answerCallback(cb.ID, "✅ Verification successful!", true)

		// Edit message to show verification successful
		successMsg := `<b>✅ Verification Successful!</b>

Welcome! You have been verified and can now use the bot.

You can get a phone number by clicking the "Get a Phone Number ☎️" button.`

		b.safeEdit(chatID, msgID, successMsg, nil)

		// Send main keyboard
		b.sendWithReplyKeyboard(chatID, "What would you like to do?", mainKeyboard())
		return
	}

	// User is not verified - show an alert pop-up (showAlert=true)
	b.answerCallback(cb.ID, "❌ You have not joined both groups yet. Please join them first.", true)
}

package bot

import (
	"fmt"
	"shark_bot/pkg/logger"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// verificationKeyboard returns the verification inline keyboard with join buttons
func (b *Bot) verificationKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📱 Join Group 1", b.verifyGroup1),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("⛧ Join Group 2", b.verifyGroup2),
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

<b>📱 Group 1:</b> SHARK SMS BACKUP
<b>⛧ Group 2:</b> SHARK METHOD

Join both groups using the buttons below, then click "Check Verification" to verify your membership.`

	b.sendHTMLWithMarkup(chatID, msg, b.verificationKeyboard())
}

// isUserVerified checks if a user has joined both required groups
// Note: For this to work, the bot must be an admin in both groups with permission to check members
func (b *Bot) isUserVerified(userID int64) bool {
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
// The group can be identified by URL (e.g., "https://t.me/SHARK_EMPIRE_1") or username/ID
func (b *Bot) checkGroupMembership(userID int64, groupIdentifier string) bool {
	if groupIdentifier == "" {
		// If group identifier is not set, skip verification
		return true
	}

	// Extract username/ID from URL if it's a full URL
	identifier := groupIdentifier
	if strings.Contains(identifier, "t.me/") {
		// Extract the part after "t.me/"
		parts := strings.Split(identifier, "t.me/")
		if len(parts) > 1 {
			identifier = parts[1]
		}
	}

	// Remove any trailing slashes
	identifier = strings.TrimSuffix(identifier, "/")

	// Try to get user's chat member status
	var chatID int64

	// Try with username first (for channels/groups)
	if strings.HasPrefix(identifier, "@") {
		chatID = 0
	} else if strings.HasPrefix(identifier, "-") {
		// It's a negative number (group ID)
		fmt.Sscanf(identifier, "%d", &chatID)
	} else {
		// Assume it's a username without @
		identifier = "@" + identifier
		chatID = 0
	}

	cfg := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	}
	if chatID == 0 {
		cfg.SuperGroupUsername = identifier
	}

	member, err := b.api.GetChatMember(cfg)
	if err != nil {
		logger.L.Error("Failed to check group membership", "user_id", userID, "group", identifier, "err", err)
		return false
	}

	// Check if user is a member (not left or kicked)
	status := member.Status
	return status == "member" || status == "administrator" || status == "creator" || status == "restricted"
}

// handleVerificationCheck handles the verification check button callback
func (b *Bot) handleVerificationCheck(cb *tgbotapi.CallbackQuery) {
	userID := cb.From.ID
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID

	if b.isUserVerified(userID) {
		// User is verified, show success and proceed to normal flow
		b.answerCallback(cb.ID, "✅ Verification successful!", false)

		// Edit message to show verification successful
		successMsg := `<b>✅ Verification Successful!</b>

Welcome! You have been verified and can now use the bot.

You can get a phone number by clicking the "Get a Phone Number ☎️" button.`

		b.safeEdit(chatID, msgID, successMsg, nil)

		// Send main keyboard
		b.sendWithReplyKeyboard(chatID, "What would you like to do?", mainKeyboard())

		// Mark user as verified (optional: store in database if needed)
		return
	}

	// User is not verified
	b.answerCallback(cb.ID, "❌ You have not joined both groups yet. Please join them first.", true)
}

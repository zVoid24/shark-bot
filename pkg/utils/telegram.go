package utils

import (
	"shark_bot/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var log = logger.New("utils")

// SendHTML safely sends an HTML-formatted message to a Telegram chat.
func SendHTML(api *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := api.Send(msg); err != nil {
		log.Error("failed to send message", "chat_id", chatID, "err", err)
	}
}

// SendHTMLWithMarkup sends an HTML message with an inline keyboard.
func SendHTMLWithMarkup(api *tgbotapi.BotAPI, chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = markup
	if _, err := api.Send(msg); err != nil {
		log.Error("failed to send message with markup", "chat_id", chatID, "err", err)
	}
}

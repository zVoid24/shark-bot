package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"shark_bot/pkg/logger"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Custom button types to support copy_text which is missing in the tgbotapi library
type CopyTextButton struct {
	Text string `json:"text"`
}

type CustomButton struct {
	Text          string          `json:"text"`
	CallbackData  string          `json:"callback_data,omitempty"`
	URL           string          `json:"url,omitempty"`
	CopyText      *CopyTextButton `json:"copy_text,omitempty"`
	CustomEmojiID string          `json:"icon_custom_emoji_id,omitempty"`
}

type CustomMarkup struct {
	InlineKeyboard [][]CustomButton `json:"inline_keyboard"`
}

var botLog = logger.New("bot")

// sendHTML sends an HTML-formatted message to a chat
func (b *Bot) sendHTML(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	if _, err := b.api.Send(msg); err != nil {
		botLog.Error("sendHTML failed", "chat_id", chatID, "err", err)
	}
}

// sendHTMLWithMarkup sends an HTML message with an inline keyboard
func (b *Bot) sendHTMLWithMarkup(chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = markup
	if _, err := b.api.Send(msg); err != nil {
		botLog.Error("sendHTMLWithMarkup failed", "chat_id", chatID, "err", err)
	}
}

// sendWithReplyKeyboard sends a message with a reply keyboard
func (b *Bot) sendWithReplyKeyboard(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		botLog.Error("sendWithReplyKeyboard failed", "chat_id", chatID, "err", err)
	}
}

// removeKeyboard sends a message removing the reply keyboard
func (b *Bot) removeKeyboard(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err := b.api.Send(msg); err != nil {
		botLog.Error("removeKeyboard failed", "chat_id", chatID, "err", err)
	}
}

// safeEdit edits an existing message, ignoring "message is not modified" errors
func (b *Bot) safeEdit(chatID int64, messageID int, text string, markup *tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = tgbotapi.ModeHTML
	edit.DisableWebPagePreview = true
	if markup != nil {
		edit.ReplyMarkup = markup
	}
	if _, err := b.api.Send(edit); err != nil {
		errStr := err.Error()
		if errStr == "Bad Request: message is not modified" ||
			errStr == "Message is not modified" {
			return // expected, ignore
		}
		botLog.Error("safeEdit failed", "chat_id", chatID, "msg_id", messageID, "err", err)
	}
}

// answerCallback answers an inline button callback to stop Telegram's loading indicator
func (b *Bot) answerCallback(callbackID, text string, showAlert bool) {
	cb := tgbotapi.NewCallback(callbackID, text)
	cb.ShowAlert = showAlert
	if _, err := b.api.Request(cb); err != nil {
		botLog.Error("answerCallback failed", "err", err)
	}
}

// mainKeyboard returns the main reply keyboard buttons
func mainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Get a Phone Number ☎️")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("📊 My Status")),
	)
}

// sendHTMLCustom sends an HTML message with custom JSON markup (for copy_text support)
// This uses a direct HTTP call to bypass library limitations on custom button types.
func (b *Bot) sendHTMLCustom(chatID int64, msgID int, text string, markup CustomMarkup) {
	method := "sendMessage"
	params := url.Values{}
	params.Set("chat_id", fmt.Sprintf("%d", chatID))
	params.Set("text", text)
	params.Set("parse_mode", "HTML")
	params.Set("disable_web_page_preview", "true")

	if msgID != 0 {
		method = "editMessageText"
		params.Set("message_id", fmt.Sprintf("%d", msgID))
	}

	markupBytes, _ := json.Marshal(markup)
	params.Set("reply_markup", string(markupBytes))

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.api.Token, method)
	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		botLog.Error("direct API request failed", "method", method, "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		botLog.Error("direct API request non-OK", "method", method, "status", resp.Status)
	}
}

// capitalize returns a string with the first letter in upper case
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[0:1]) + strings.ToLower(s[1:])
}

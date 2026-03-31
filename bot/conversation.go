package bot

import (
	"fmt"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Conversation states
const (
	convStepNone           = 0
	convStepChoosePlat     = 2
	convStepNewPlatName    = 3
	convStepChooseCountry  = 4
	convStepNewCountryName = 5
	convStepAwaitFile      = 6
	convStepRemovePlat     = 7
	convStepRemoveCountry  = 8
)

type convContext struct {
	Step     int
	Platform string
	Country  string
}

var convMu sync.Mutex

func (b *Bot) setConvState(userID int64, ctx *convContext) {
	convMu.Lock()
	defer convMu.Unlock()
	if ctx == nil {
		delete(b.convState, userID)
	} else {
		b.convState[userID] = ctx
	}
}

func (b *Bot) getConvState(userID int64) *convContext {
	convMu.Lock()
	defer convMu.Unlock()
	return b.convState[userID]
}

// handleConversationText handles text input during a conversation flow.
// Returns true if the update was handled.
func (b *Bot) handleConversationText(msg *tgbotapi.Message) bool {
	ctx := b.getConvState(msg.From.ID)
	if ctx == nil {
		return false
	}

	switch ctx.Step {

	// --- Add Number: Choose Platform ---
	case convStepChoosePlat:
		if msg.Text == "➕ New Platform" {
			b.setConvState(msg.From.ID, &convContext{Step: convStepNewPlatName})
			b.removeKeyboard(msg.Chat.ID, "<b>Enter the new platform name (e.g., WhatsApp):</b>")
		} else {
			plat := msg.Text
			ctx.Platform = plat
			ctx.Step = convStepChooseCountry
			b.setConvState(msg.From.ID, ctx)
			b.showCountryKeyboard(msg.Chat.ID, plat)
		}
		return true

	// --- Add Number: New Platform Name ---
	case convStepNewPlatName:
		ctx.Platform = strings.TrimSpace(msg.Text)
		ctx.Step = convStepNewCountryName
		b.setConvState(msg.From.ID, ctx)
		b.removeKeyboard(msg.Chat.ID, "<b>Enter country name (with flag):</b>")
		return true

	// --- Add Number: Choose Country ---
	case convStepChooseCountry:
		if msg.Text == "➕ New Country" {
			ctx.Step = convStepNewCountryName
			b.setConvState(msg.From.ID, ctx)
			b.removeKeyboard(msg.Chat.ID, "<b>Enter new country name (with flag):</b>")
		} else {
			ctx.Country = msg.Text
			ctx.Step = convStepAwaitFile
			b.setConvState(msg.From.ID, ctx)
			b.removeKeyboard(msg.Chat.ID, fmt.Sprintf("<b>Selected: %s - %s</b>\nUpload .txt file:", ctx.Platform, ctx.Country))
		}
		return true

	// --- Add Number: New Country Name ---
	case convStepNewCountryName:
		ctx.Country = strings.TrimSpace(msg.Text)
		ctx.Step = convStepAwaitFile
		b.setConvState(msg.From.ID, ctx)
		b.removeKeyboard(msg.Chat.ID, fmt.Sprintf("<b>Adding numbers for: %s - %s</b>\nUpload .txt file:", ctx.Platform, ctx.Country))
		return true

	// --- Remove Number: Choose Platform ---
	case convStepRemovePlat:
		ctx.Platform = msg.Text
		ctx.Step = convStepRemoveCountry
		b.setConvState(msg.From.ID, ctx)
		b.showRemoveCountryKeyboard(msg.Chat.ID, msg.Text)
		return true

	// --- Remove Number: Choose Country ---
	case convStepRemoveCountry:
		plat := ctx.Platform
		coun := msg.Text
		_ = b.numberSvc.DeleteByPlatformCountry(plat, coun)
		b.setConvState(msg.From.ID, nil)
		kb := tgbotapi.NewRemoveKeyboard(true)
		m := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("<b>%s has been successfully removed from %s.</b>", coun, plat))
		m.ParseMode = tgbotapi.ModeHTML
		m.ReplyMarkup = kb
		b.api.Send(m)
		return true
	}

	return false
}

// handleConversationDocument handles file uploads during conversation
func (b *Bot) handleConversationDocument(msg *tgbotapi.Message) bool {
	ctx := b.getConvState(msg.From.ID)
	if ctx == nil || ctx.Step != convStepAwaitFile {
		return false
	}

	plat := ctx.Platform
	coun := ctx.Country

	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: msg.Document.FileID})
	if err != nil {
		b.sendHTML(msg.Chat.ID, "<b>Error getting file.</b>")
		return true
	}

	content, err := downloadFile(b.api, file)
	if err != nil {
		b.sendHTML(msg.Chat.ID, "<b>Error downloading file.</b>")
		return true
	}

	lines := strings.Split(string(content), "\n")
	count, _ := b.numberSvc.BulkInsert(plat, coun, lines)
	b.setConvState(msg.From.ID, nil)

	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>Successfully added %d numbers to %s (%s).</b>", count, plat, coun))

	// Broadcast new numbers notification
	broadcastMsg := fmt.Sprintf(
		"<b>🚀 New Numbers Added!</b>\n\n<b>Platform:</b> <code>%s</code>\n<b>Country:</b> <code>%s</code>\n<b>Quantity:</b> <code>%d</code> <b>numbers</b>\n\n<i>Get your number now using the button below!</i>",
		plat, coun, count)
	userIDs, _ := b.userSvc.GetUnblockedUserIDs()
	go b.runBroadcast(userIDs, broadcastMsg, msg.Chat.ID)

	return true
}

// handleAddNumber starts add number conversation
func (b *Bot) handleAddNumber(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		b.sendHTML(msg.Chat.ID, "<b>Sorry, this command is for the owner only.</b>")
		return
	}
	platforms, _ := b.numberSvc.GetPlatforms()
	var rows [][]tgbotapi.KeyboardButton
	for _, p := range platforms {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(p)))
	}
	rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("➕ New Platform")))
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.OneTimeKeyboard = true
	kb.ResizeKeyboard = true
	b.setConvState(msg.From.ID, &convContext{Step: convStepChoosePlat})
	m := tgbotapi.NewMessage(msg.Chat.ID, "<b>Select a platform to add numbers to:</b>")
	m.ParseMode = tgbotapi.ModeHTML
	m.ReplyMarkup = kb
	b.api.Send(m)
}

// handleRemoveNumber starts remove number conversation
func (b *Bot) handleRemoveNumber(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		b.sendHTML(msg.Chat.ID, "<b>Sorry, this command is for the owner only.</b>")
		return
	}
	platforms, _ := b.numberSvc.GetPlatforms()
	if len(platforms) == 0 {
		b.sendHTML(msg.Chat.ID, "<b>There are no platforms to remove.</b>")
		return
	}
	var rows [][]tgbotapi.KeyboardButton
	for _, p := range platforms {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(p)))
	}
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.OneTimeKeyboard = true
	b.setConvState(msg.From.ID, &convContext{Step: convStepRemovePlat})
	m := tgbotapi.NewMessage(msg.Chat.ID, "<b>Select the platform:</b>")
	m.ParseMode = tgbotapi.ModeHTML
	m.ReplyMarkup = kb
	b.api.Send(m)
}

// handleCancel cancels current conversation
func (b *Bot) handleCancel(msg *tgbotapi.Message) {
	b.setConvState(msg.From.ID, nil)
	m := tgbotapi.NewMessage(msg.Chat.ID, "<b>Operation cancelled.</b>")
	m.ParseMode = tgbotapi.ModeHTML
	m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	b.api.Send(m)
}

// showCountryKeyboard shows reply keyboard of countries for platform
func (b *Bot) showCountryKeyboard(chatID int64, platform string) {
	countries, _ := b.numberSvc.GetCountries(platform)
	var rows [][]tgbotapi.KeyboardButton
	for _, c := range countries {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(c)))
	}
	rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("➕ New Country")))
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.OneTimeKeyboard = true
	kb.ResizeKeyboard = true
	m := tgbotapi.NewMessage(chatID, fmt.Sprintf("<b>Platform: %s</b>\nSelect a country or add new:", platform))
	m.ParseMode = tgbotapi.ModeHTML
	m.ReplyMarkup = kb
	b.api.Send(m)
}

// showRemoveCountryKeyboard shows reply keyboard for removing a country
func (b *Bot) showRemoveCountryKeyboard(chatID int64, platform string) {
	countries, _ := b.numberSvc.GetCountries(platform)
	var rows [][]tgbotapi.KeyboardButton
	for _, c := range countries {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(c)))
	}
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.OneTimeKeyboard = true
	m := tgbotapi.NewMessage(chatID, "<b>Select the country to remove:</b>")
	m.ParseMode = tgbotapi.ModeHTML
	m.ReplyMarkup = kb
	b.api.Send(m)
}

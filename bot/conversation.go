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
	convStepAwaitPlatform  = 9
)

type convContext struct {
	Step      int
	Platform  string
	Country   string
	Platforms []string // Selected platforms for upload
	Lines     []string // Numbers from file
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
			plat := strings.ToLower(msg.Text)
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
			b.removeKeyboard(msg.Chat.ID, fmt.Sprintf("<b>Selected: %s - %s</b>\nUpload .txt file:", capitalize(ctx.Platform), ctx.Country))
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
	userID := msg.From.ID
	ctx := b.getConvState(userID)

	isAdmin, _ := b.adminSvc.IsAdmin(fmt.Sprintf("%d", userID))
	if !isAdmin {
		return false
	}

	// Case 1: Manual flow (deprecated but kept for compatibility if needed)
	if ctx != nil && ctx.Step == convStepAwaitFile {
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

	// Case 2: New direct upload flow with auto-country detection
	if strings.HasSuffix(strings.ToLower(msg.Document.FileName), ".txt") {
		// Detect country from filename: CountryName_Numbers.txt
		parts := strings.Split(msg.Document.FileName, "_")
		if len(parts) < 2 {
			// If it doesn't match the pattern, don't auto-handle unless in a state
			if ctx == nil {
				return false
			}
		}

		country := parts[0]
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
		var cleanLines []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				cleanLines = append(cleanLines, l)
			}
		}

		if len(cleanLines) == 0 {
			b.sendHTML(msg.Chat.ID, "<b>The uploaded file is empty.</b>")
			return true
		}

		// Set state to await platform selection
		b.setConvState(userID, &convContext{
			Step:     convStepAwaitPlatform,
			Country:  country,
			Lines:    cleanLines,
			Platforms: []string{},
		})

		b.showUploadPlatformSelector(msg.Chat.ID, country, []string{})
		return true
	}

	return false
}

func (b *Bot) showUploadPlatformSelector(chatID int64, country string, selected []string) {
	platforms := []string{"facebook", "instagram", "whatsapp", "imo", "telegram"}
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, p := range platforms {
		isSelected := false
		for _, s := range selected {
			if s == p {
				isSelected = true
				break
			}
		}

		label := p
		if isSelected {
			label = "✅ " + p
		}

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "toggle_upload_plat::"+p),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Done", "confirm_upload"),
	))

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	text := fmt.Sprintf("<b>Upload Numbers</b>\n\n<b>Country:</b> <code>%s</code>\nSelect platforms for these numbers:", country)

	// Check if message already exists to edit or send new
	// In conversation, we usually send new. But for toggling we'll edit.
	// For the initial show, we send.
	b.sendHTMLWithMarkup(chatID, text, markup)
}

func (b *Bot) handleUploadPlatToggle(cb *tgbotapi.CallbackQuery, plat string) {
	ctx := b.getConvState(cb.From.ID)
	if ctx == nil || ctx.Step != convStepAwaitPlatform {
		return
	}

	found := false
	for i, p := range ctx.Platforms {
		if p == plat {
			ctx.Platforms = append(ctx.Platforms[:i], ctx.Platforms[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		ctx.Platforms = append(ctx.Platforms, plat)
	}

	b.setConvState(cb.From.ID, ctx)

	// Update the keyboard
	platforms := []string{"facebook", "instagram", "whatsapp", "imo", "telegram"}
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, p := range platforms {
		isSelected := false
		for _, s := range ctx.Platforms {
			if s == p {
				isSelected = true
				break
			}
		}

		label := p
		if isSelected {
			label = "✅ " + p
		}

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "toggle_upload_plat::"+p),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Done", "confirm_upload"),
	))

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	text := fmt.Sprintf("<b>Upload Numbers</b>\n\n<b>Country:</b> <code>%s</code>\nSelect platforms for these numbers:", ctx.Country)

	b.safeEdit(cb.Message.Chat.ID, cb.Message.MessageID, text, &markup)
}

func (b *Bot) handleUploadConfirm(cb *tgbotapi.CallbackQuery) {
	userID := cb.From.ID
	ctx := b.getConvState(userID)
	if ctx == nil || ctx.Step != convStepAwaitPlatform {
		return
	}

	if len(ctx.Platforms) == 0 {
		b.answerCallback(cb.ID, "❌ Please select at least one platform.", true)
		return
	}

	totalInserted := 0
	platformsStr := strings.Join(ctx.Platforms, ", ")

	for _, plat := range ctx.Platforms {
		count, _ := b.numberSvc.BulkInsert(plat, ctx.Country, ctx.Lines)
		totalInserted += count
	}

	b.setConvState(userID, nil)

	// Notify admin
	b.safeEdit(cb.Message.Chat.ID, cb.Message.MessageID,
		fmt.Sprintf("<b>✅ Upload Successful!</b>\n\n<b>Country:</b> <code>%s</code>\n<b>Platforms:</b> <code>%s</code>\n<b>Total Numbers Added:</b> <code>%d</code>",
			ctx.Country, platformsStr, totalInserted),
		nil)

	// Broadcast
	broadcastMsg := fmt.Sprintf(
		"<b>🚀 New Numbers Added!</b>\n\n<b>Platform:</b> <code>%s</code>\n<b>Country:</b> <code>%s</code>\n<b>Quantity:</b> <code>%d</code> <b>numbers</b>\n\n<i>Get your number now using the button below!</i>",
		platformsStr, ctx.Country, totalInserted)
	userIDs, _ := b.userSvc.GetUnblockedUserIDs()
	go b.runBroadcast(userIDs, broadcastMsg, cb.Message.Chat.ID)
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
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(capitalize(p))))
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
	m := tgbotapi.NewMessage(chatID, fmt.Sprintf("<b>Platform: %s</b>\nSelect a country or add new:", capitalize(platform)))
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

package bot

import (
	"context"
	"fmt"
	"shark_bot/internal/activenumber"
	"shark_bot/pkg/logger"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// helper: make keyboard rows with N buttons per row
func buildButtonRows(buttons []tgbotapi.InlineKeyboardButton, perRow int) [][]tgbotapi.InlineKeyboardButton {
	if perRow <= 0 {
		perRow = 1
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(buttons); i += perRow {
		end := i + perRow
		if end > len(buttons) {
			end = len(buttons)
		}
		rows = append(rows, buttons[i:end])
	}
	return rows
}

// handleGetNumber shows the platform list
func (b *Bot) handleGetNumber(msg *tgbotapi.Message) {
	if blocked, _ := b.userSvc.IsBlocked(fmt.Sprintf("%d", msg.From.ID)); blocked {
		return
	}
	if err := b.showPlatformList(msg.Chat.ID, msg.MessageID, false, fmt.Sprintf("%d", msg.From.ID)); err != nil {
		logger.L.Error("handleGetNumber failed", "err", err)
	}
}

// showPlatformList shows inline keyboard of all platforms
func (b *Bot) showPlatformList(chatID int64, msgID int, isEdit bool, userID string) error {
	platforms, err := b.numberSvc.GetPlatforms()
	if err != nil {
		return err
	}

	isAdmin := b.isAdmin(userID)

	if len(platforms) == 0 {
		text := "<b>No platform available</b>"
		if isEdit {
			b.safeEdit(chatID, msgID, text, nil)
		} else {
			b.sendHTML(chatID, text)
		}
		return nil
	}

	var rows [][]CustomButton

	for _, p := range platforms {
		label := capitalize(p)
		emojiID := GetPlatformEmojiID(p)

		if isAdmin {
			count, err := b.numberSvc.CountAvailable(p, "")
			if err == nil {
				label = fmt.Sprintf("%s (%d)", label, count)
			}
		}

		rows = append(rows, []CustomButton{
			{
				Text:          label,
				CallbackData:  "select_platform::" + p,
				CustomEmojiID: emojiID,
			},
		})
	}

	markup := CustomMarkup{InlineKeyboard: rows}

	// ✅ Wider looking message
	text := fmt.Sprintf("<b>%s</b>\nChoose a service below to continue.", capitalize("platform selection"))
	if isAdmin {
		text = "<b>Select platform (Admin View)</b>\nChoose a service below to continue."
	}

	// If not an edit, we MUST send a new message (msgID=0 for sendHTMLCustom)
	targetMsgID := msgID
	if !isEdit {
		targetMsgID = 0
	}

	b.sendHTMLCustom(chatID, targetMsgID, text, markup)

	return nil
}

// showCountryList shows inline keyboard of countries for a platform
func (b *Bot) showCountryList(chatID int64, msgID int, platform, userID string) {
	countries, err := b.numberSvc.GetCountries(platform)
	if err != nil || len(countries) == 0 {
		b.safeEdit(chatID, msgID, "<b>No country available</b>", nil)
		return
	}

	isAdmin := b.isAdmin(userID)

	var rows [][]CustomButton

	for _, c := range countries {
		label := c
		emojiID := GetFlagEmojiIDByName(c)

		if emojiID == "" {
			label = GetFlagEmojiByName(c) + " " + c
		}

		if isAdmin {
			count, err := b.numberSvc.CountAvailable(platform, c)
			if err == nil {
				label = fmt.Sprintf("%s (%d)", label, count)
			}
		}

		rows = append(rows, []CustomButton{
			{
				Text:          label,
				CallbackData:  fmt.Sprintf("select_country::%s::%s", platform, c),
				CustomEmojiID: emojiID,
			},
		})
	}

	// Group buttons into 2 per row
	var groupedRows [][]CustomButton
	for i := 0; i < len(rows); i += 2 {
		end := i + 2
		if end > len(rows) {
			end = len(rows)
		}
		var row []CustomButton
		for j := i; j < end; j++ {
			row = append(row, rows[j][0])
		}
		groupedRows = append(groupedRows, row)
	}

	groupedRows = append(groupedRows, []CustomButton{
		{
			Text:         "⬅ Back",
			CallbackData: "back_to_platforms",
		},
	})

	markup := CustomMarkup{InlineKeyboard: groupedRows}

	text := fmt.Sprintf(
		"<b>%s</b>\nSelect a country to continue.",
		capitalize(platform),
	)

	b.sendHTMLCustom(chatID, msgID, text, markup)
}

// assignNumbers picks numbers and assigns them to the user
func (b *Bot) assignNumbers(chatID int64, userID int64, platform, country string, msgID int, isChange bool) {
	userIDStr := fmt.Sprintf("%d", userID)

	var excludeNums []string

	if isChange {
		// Get currently held numbers to exclude and delete
		actives, _ := b.activeSvc.GetByUser(userIDStr)
		for _, a := range actives {
			excludeNums = append(excludeNums, a.Number)
			_ = b.numberSvc.UpdateLastUsed(a.Number, a.Platform, a.Country)
			if b.activeCache != nil {
				_ = b.activeCache.DeleteByNumber(context.Background(), a.Number, a.Platform)
			}
		}

		_ = b.activeSvc.DeleteByUser(userIDStr)
		if b.activeCache != nil {
			_ = b.activeCache.DeleteByUser(context.Background(), userIDStr)
		}
	} else {
		// Delete any old numbers first
		actives, _ := b.activeSvc.GetByUser(userIDStr)
		for _, a := range actives {
			_ = b.numberSvc.UpdateLastUsed(a.Number, a.Platform, a.Country)
			if b.activeCache != nil {
				_ = b.activeCache.DeleteByNumber(context.Background(), a.Number, a.Platform)
			}
		}

		_ = b.activeSvc.DeleteByUser(userIDStr)
		if b.activeCache != nil {
			_ = b.activeCache.DeleteByUser(context.Background(), userIDStr)
		}
	}

	// Fetch up to 3 numbers
	numbers, err := b.numberSvc.GetNumbers(platform, country, userIDStr, excludeNums, 3)
	if err != nil || len(numbers) == 0 {
		b.safeEdit(chatID, msgID,
			fmt.Sprintf("<b>❌ No number available</b>\n%s • %s", platform, country),
			nil,
		)
		return
	}

	var rows [][]CustomButton
	var copyRow []CustomButton

	numbersText := ""
	for i, num := range numbers {
		an := activenumber.ActiveNumber{
			Number:    num,
			UserID:    userIDStr,
			Timestamp: time.Now(),
			MessageID: int64(msgID),
			Platform:  platform,
			Country:   country,
		}
		_ = b.activeSvc.Insert(an)
		if b.activeCache != nil {
			_ = b.activeCache.Set(context.Background(), an)
		}
		_ = b.seenSvc.Add(userIDStr, num, country)

		// Add number to text (monospaced)
		numbersText += fmt.Sprintf("📱 <b>#%d:</b> <code>%s</code>\n", i+1, num)

		// Add to copy row
		copyRow = append(copyRow, CustomButton{
			Text:          fmt.Sprintf("#%d", i+1),
			CopyText:      &CopyTextButton{Text: num},
			CustomEmojiID: "6176966310920983412",
		})
	}

	// Add the grid of copy buttons
	if len(copyRow) > 0 {
		rows = append(rows, copyRow)
	}

	_ = b.activeSvc.UpdateMessageID(userIDStr, int64(msgID))

	text := fmt.Sprintf(
		"<b>%s %s - %s %s</b>\n"+
			"────────────────────────\n"+
			"%s\n"+
			"<i>⏳ Waiting for OTP... (Auto-expiry: 1h)</i>",
		GetPlatformEmoji(platform), capitalize(platform), GetFlagByName(country), country, numbersText,
	)

	// Add navigation buttons
	rows = append(rows, []CustomButton{
		{
			Text:          "Change (10s CD)",
			CallbackData:  fmt.Sprintf("change_number::%s::%s", platform, country),
			CustomEmojiID: "5231197925178089666",
		},
		{
			Text:          "Back",
			CallbackData:  fmt.Sprintf("back_to_countries::%s", platform),
			CustomEmojiID: "5471949924658588235", // Telegram-style arrow
		},
	})
	rows = append(rows, []CustomButton{
		{
			Text:          "OTP Group",
			URL:           "https://t.me/shark_sms_panel",
			CustomEmojiID: "5330237710655306682", // Telegram logo
		},
		{
			Text:          "Guide",
			URL:           "https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz",
			CustomEmojiID: "5942902988564600402", // Method/Guide icon
		},
	})

	markup := CustomMarkup{InlineKeyboard: rows}

	// Because tgbotapi.EditMessageTextConfig expects an InlineKeyboardMarkup,
	// and our CustomMarkup is a raw struct with JSON tags, we need to send it
	// via a manual request if we want to use 'copy_text'.
	// For now, I'll use a helper to send it.

	b.sendHTMLCustom(chatID, msgID, text, markup)
}

// handleMyStatus shows the user's OTP usage stats
func (b *Bot) handleMyStatus(msg *tgbotapi.Message) {
	userID := fmt.Sprintf("%d", msg.From.ID)
	stats, err := b.statsSvc.GetUserOtpStats(userID)
	if err != nil || len(stats) == 0 {
		b.sendHTML(msg.Chat.ID, "<b>No OTP status available yet.</b>")
		return
	}

	text := "<b>📊 My OTP Usage</b>\n\n"
	total := 0
	for _, s := range stats {
		text += fmt.Sprintf("%s · <code>%d</code>\n", s.Country, s.Count)
		total += s.Count
	}
	text += fmt.Sprintf("\n<b>Total</b> · <code>%d</code>", total)

	b.sendHTML(msg.Chat.ID, text)
}

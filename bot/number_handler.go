package bot

import (
	"context"
	"fmt"
	"shark_bot/pkg/logger"
	"time"

	"shark_bot/internal/activenumber"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleGetNumber shows the platform list
func (b *Bot) handleGetNumber(msg *tgbotapi.Message) {
	if blocked, _ := b.userSvc.IsBlocked(fmt.Sprintf("%d", msg.From.ID)); blocked {
		return
	}
	if err := b.showPlatformList(msg.Chat.ID, msg.MessageID, false); err != nil {
		logger.L.Error("handleGetNumber failed", "err", err)
	}
}

// showPlatformList shows inline keyboard of all platforms
func (b *Bot) showPlatformList(chatID int64, msgID int, isEdit bool) error {
	platforms, err := b.numberSvc.GetPlatforms()
	if err != nil {
		return err
	}

	if len(platforms) == 0 {
		text := "<b>Sorry, no platforms are available right now.</b>"
		if isEdit {
			b.safeEdit(chatID, msgID, text, nil)
		} else {
			b.sendHTML(chatID, text)
		}
		return nil
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, p := range platforms {
		count, _ := b.numberSvc.CountAvailable(p, "")
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s (%d)", p, count),
			"select_platform::"+p,
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	text := "<b>🔧 Select the platform you need to access:</b>"

	if isEdit {
		b.safeEdit(chatID, msgID, text, &markup)
	} else {
		b.sendHTMLWithMarkup(chatID, text, markup)
	}
	return nil
}

// showCountryList shows inline keyboard of countries for a platform
func (b *Bot) showCountryList(chatID int64, msgID int, platform string) {
	countries, err := b.numberSvc.GetCountries(platform)
	if err != nil || len(countries) == 0 {
		b.safeEdit(chatID, msgID, fmt.Sprintf("<b>Sorry, no countries available for %s.</b>", platform), nil)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, c := range countries {
		count, _ := b.numberSvc.CountAvailable(platform, c)
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s (%d)", c, count),
			fmt.Sprintf("select_country::%s::%s", platform, c),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	// Back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Back to Platforms", "back_to_platforms"),
	))
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.safeEdit(chatID, msgID, fmt.Sprintf("<b>Select your country for %s:</b>", platform), &markup)
}

// assignNumbers picks numbers and assigns them to the user
func (b *Bot) assignNumbers(chatID int64, userID int64, platform, country string, msgID int, isChange bool) {
	userIDStr := fmt.Sprintf("%d", userID)

	var excludeNums []string

	if isChange {
		// Get currently held numbers to exclude
		actives, _ := b.activeSvc.GetByUser(userIDStr)
		for _, a := range actives {
			excludeNums = append(excludeNums, a.Number)
			// Permanently delete the old number so no one else gets it
			_ = b.numberSvc.DeleteByNumber(a.Number)
			if b.activeCache != nil {
				_ = b.activeCache.DeleteByNumber(context.Background(), a.Number)
			}
		}
		// Release all active numbers for user
		_ = b.activeSvc.DeleteByUser(userIDStr)
		if b.activeCache != nil {
			_ = b.activeCache.DeleteByUser(context.Background(), userIDStr)
		}
	} else {
		// Permanently delete any old ones first
		actives, _ := b.activeSvc.GetByUser(userIDStr)
		for _, a := range actives {
			_ = b.numberSvc.DeleteByNumber(a.Number)
			if b.activeCache != nil {
				_ = b.activeCache.DeleteByNumber(context.Background(), a.Number)
			}
		}
		// Release from active list
		_ = b.activeSvc.DeleteByUser(userIDStr)
		if b.activeCache != nil {
			_ = b.activeCache.DeleteByUser(context.Background(), userIDStr)
		}
	}

	// limit := b.settingsSvc.GetNumberLimit(platform, country)
	numbers, err := b.numberSvc.GetNumbers(platform, country, userIDStr, excludeNums, 1)
	if err != nil || len(numbers) == 0 {
		b.safeEdit(chatID, msgID,
			fmt.Sprintf("<b>Sorry, no numbers are currently available for %s.</b>", country), nil)
		return
	}

	// Insert into active_numbers
	for _, num := range numbers {
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
	}

	// Update message_id for all newly assigned numbers (in case Insert used a stale msgID)
	_ = b.activeSvc.UpdateMessageID(userIDStr, int64(msgID))

	numDisplay := ""
	for _, n := range numbers {
		numDisplay += fmt.Sprintf("<code>%s</code>\n", n)
	}

	text := fmt.Sprintf(
	"<b>%s (%s) Number(s) Assigned:</b>\n%s\n<b>Waiting for OTP...</b>\n\nDeveloped by <a href=\"https://t.me/zVoid24\">𝑍𝐴𝐻𝐼𝐷</a>\n\n<b>📺 Method Guide</b>\n<a href=\"https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz\">Click to Watch</a>",
	country, platform, numDisplay,
)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Change Number",
				fmt.Sprintf("change_number::%s::%s", platform, country)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📢 OTP Group", "https://t.me/shark_sms_panel"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back",
				fmt.Sprintf("back_to_countries::%s", platform)),
		),
	)

	b.safeEdit(chatID, msgID, text, &markup)
}

// handleMyStatus shows the user's OTP usage stats
func (b *Bot) handleMyStatus(msg *tgbotapi.Message) {
	userID := fmt.Sprintf("%d", msg.From.ID)
	stats, err := b.statsSvc.GetUserOtpStats(userID)
	if err != nil || len(stats) == 0 {
		b.sendHTML(msg.Chat.ID, "<b>No OTP status available yet.</b>")
		return
	}

	text := "<b>📊 My OTP Usage 📊</b>\n\n"
	total := 0
	for _, s := range stats {
		text += fmt.Sprintf("<b>%s:</b> <code>%d</code> <b>OTPs</b>\n", s.Country, s.Count)
		total += s.Count
	}
	text += fmt.Sprintf("\n<b>Total:</b> <code>%d</code>", total)
	b.sendHTML(msg.Chat.ID, text)
}

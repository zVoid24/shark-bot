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
		text := "<b>No platform available</b>"
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

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s (%d)", p, count),
				"select_platform::"+p,
			),
		))
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	// ✅ Wider looking message
	text := "<b>Select platform</b>\nChoose a service below to continue."

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
		b.safeEdit(chatID, msgID, "<b>No country available</b>", nil)
		return
	}

	var buttons []tgbotapi.InlineKeyboardButton

	for _, c := range countries {
		count, _ := b.numberSvc.CountAvailable(platform, c)

		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s (%d)", c, count),
				fmt.Sprintf("select_country::%s::%s", platform, c),
			),
		)
	}

	rows := buildButtonRows(buttons, 2)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅ Back", "back_to_platforms"),
	))

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	// ✅ Make message feel wider
	text := fmt.Sprintf(
		"<b>%s</b>\nSelect a country to continue.",
		platform,
	)

	b.safeEdit(chatID, msgID, text, &markup)
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
			_ = b.numberSvc.DeleteByNumber(a.Number)
			if b.activeCache != nil {
				_ = b.activeCache.DeleteByNumber(context.Background(), a.Number)
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
			_ = b.numberSvc.DeleteByNumber(a.Number)
			if b.activeCache != nil {
				_ = b.activeCache.DeleteByNumber(context.Background(), a.Number)
			}
		}

		_ = b.activeSvc.DeleteByUser(userIDStr)
		if b.activeCache != nil {
			_ = b.activeCache.DeleteByUser(context.Background(), userIDStr)
		}
	}

	numbers, err := b.numberSvc.GetNumbers(platform, country, userIDStr, excludeNums, 1)
	if err != nil || len(numbers) == 0 {
		b.safeEdit(chatID, msgID,
			fmt.Sprintf("<b>No number available</b>\n%s • %s", platform, country),
			nil,
		)
		return
	}

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

	_ = b.activeSvc.UpdateMessageID(userIDStr, int64(msgID))

	number := numbers[0]

	text := fmt.Sprintf(
    "<b>✅ Number assigned</b>\n\n<code>%s</code>\n\n<b>Platform:</b> %s\n<b>Country:</b> %s\n\n<i>Waiting for OTP...</i>",
    number, platform, country,
)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔄 Change",
				fmt.Sprintf("change_number::%s::%s", platform, country),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"⬅ Back",
				fmt.Sprintf("back_to_countries::%s", platform),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(
				"📢 OTP Group",
				"https://t.me/shark_sms_panel",
			),
			tgbotapi.NewInlineKeyboardButtonURL(
				"📺 Guide",
				"https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz",
			),
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

	text := "<b>📊 My OTP Usage</b>\n\n"
	total := 0
	for _, s := range stats {
		text += fmt.Sprintf("%s · <code>%d</code>\n", s.Country, s.Count)
		total += s.Count
	}
	text += fmt.Sprintf("\n<b>Total</b> · <code>%d</code>", total)

	b.sendHTML(msg.Chat.ID, text)
}
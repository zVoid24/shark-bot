package bot

import (
	"context"
	"fmt"
	"shark_bot/internal/activenumber"
	"shark_bot/pkg/logger"
	"strconv"
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

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, p := range platforms {
		label := p
		if isAdmin {
			count, err := b.numberSvc.CountAvailable(p, "")
			if err == nil {
				label = fmt.Sprintf("%s (%d)", p, count)
			}
		}

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				label,
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
func (b *Bot) showCountryList(chatID int64, msgID int, platform, userID string) {
	countries, err := b.numberSvc.GetCountries(platform)
	if err != nil || len(countries) == 0 {
		b.safeEdit(chatID, msgID, "<b>No country available</b>", nil)
		return
	}

	isAdmin := b.isAdmin(userID)

	var buttons []tgbotapi.InlineKeyboardButton

	for _, c := range countries {
		label := c
		if isAdmin {
			count, err := b.numberSvc.CountAvailable(platform, c)
			if err == nil {
				label = fmt.Sprintf("%s (%d)", c, count)
			}
		}

		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(
				label,
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
			_ = b.numberSvc.DeleteSpecific(a.Number, a.Platform, a.Country)
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
			_ = b.numberSvc.DeleteSpecific(a.Number, a.Platform, a.Country)
			if b.activeCache != nil {
				_ = b.activeCache.DeleteByNumber(context.Background(), a.Number, a.Platform)
			}
		}

		_ = b.activeSvc.DeleteByUser(userIDStr)
		if b.activeCache != nil {
			_ = b.activeCache.DeleteByUser(context.Background(), userIDStr)
		}
	}

	numbers, err := b.numberSvc.GetNumbers(platform, country, userIDStr, excludeNums, 3)
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

	numbersStr := ""
	for i, num := range numbers {
		numbersStr += fmt.Sprintf("<b>#%d:</b> <code>%s</code>\n", i+1, num)
	}

	text := fmt.Sprintf(
		`<tg-emoji emoji-id="5469931148295547357">✅</tg-emoji><b> Numbers assigned successfully</b>

%s
<b>Platform:</b> %s
<b>Country:</b> %s

<i>Waiting for OTP... (Auto-expiry: 1 hour)</i>`,
		numbersStr, platform, country,
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

// handleMyStatus shows the user's OTP usage stats and total earnings
func (b *Bot) handleMyStatus(msg *tgbotapi.Message) {
	userID := fmt.Sprintf("%d", msg.From.ID)
	
	// Fetch User info for balance
	u, _ := b.userSvc.GetUser(userID)
	balance := 0.0
	if u != nil {
		balance = u.Balance
	}

	stats, err := b.statsSvc.GetUserOtpStats(userID)
	if err != nil {
		b.sendHTML(msg.Chat.ID, "<b>An error occurred fetching stats.</b>")
		return
	}

	text := "<b>📊 My Account Status</b>\n\n"
	text += fmt.Sprintf("<b>💰 Total Earnings:</b> <code>$%s</code>\n\n", strconv.FormatFloat(balance, 'f', -1, 64))
	
	if len(stats) == 0 {
		text += "<b>No OTP activity yet.</b>"
	} else {
		text += "<b>--- OTP Breakdown ---</b>\n"
		total := 0
		for _, s := range stats {
			text += fmt.Sprintf("%s · <code>%d</code>\n", s.Country, s.Count)
			total += s.Count
		}
		text += fmt.Sprintf("\n<b>Total OTPs:</b> <code>%d</code>", total)
	}

	b.sendHTML(msg.Chat.ID, text)
}

func (b *Bot) handleWallet(msg *tgbotapi.Message) {
	userID := fmt.Sprintf("%d", msg.From.ID)
	u, _ := b.userSvc.GetUser(userID)
	balance := 0.0
	if u != nil {
		balance = u.Balance
	}
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>💰 Your Current Balance</b>\n\n<b>Balance:</b> <code>$%s</code>", strconv.FormatFloat(balance, 'f', -1, 64)))
}

func (b *Bot) handleWithdrawStart(msg *tgbotapi.Message) {
	userID := fmt.Sprintf("%d", msg.From.ID)
	u, _ := b.userSvc.GetUser(userID)
	balance := 0.0
	if u != nil {
		balance = u.Balance
	}

	minWithdrawStr, _ := b.settingsSvc.Get("min_withdraw")
	minWithdraw := 0.0
	if minWithdrawStr != "" {
		fmt.Sscanf(minWithdrawStr, "%f", &minWithdraw)
	}

	if balance < minWithdraw {
		b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>❌ Withdrawal Denied</b>\n\nMinimum withdrawal amount is <code>$%s</code>.\nYour current balance: <code>$%s</code>", strconv.FormatFloat(minWithdraw, 'f', -1, 64), strconv.FormatFloat(balance, 'f', -1, 64)))
		return
	}

	// Initiate conversation flow for Binance ID
	b.setConvState(msg.From.ID, &convContext{Step: convStepAwaitBinanceID})
	b.removeKeyboard(msg.Chat.ID, "<b>💳 Withdrawal Request</b>\n\nPlease enter your <b>Binance ID</b> to proceed:")
}

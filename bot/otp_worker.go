package bot

import (
	"database/sql"
	"fmt"
	"regexp"
	"shark_bot/pkg/logger"
	"strings"
	"time"

	"shark_bot/internal/activenumber"
	"shark_bot/internal/processednumber"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// OTP regex patterns matching the Python version exactly
var otpPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:Your Code|Code|OTP|Codigo|verification|OTP Code)\s*(?:➡️|:|\s)\s*([\d\s-]+)`),
	regexp.MustCompile(`(?i)G-([\d]+) is your Google verification code`),
	regexp.MustCompile(`(?i)#\s*([\d]+)\s*is your Facebook code`),
	regexp.MustCompile(`(?i)Your WhatsApp(?:\s+Business)? code\s*([\d\s-]+)`),
	regexp.MustCompile(`\b(\d{3}[-\s]\d{3,4})\b`),
	regexp.MustCompile(`(?i)code is\s*[:\s]*(\d{4,8})`),
	regexp.MustCompile(`(?i)code:\s*(\d{4,8})`),
	regexp.MustCompile(`\b(\d{4,8})\b`),
}

var numberPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:Number|Mobile|Phone|📱|☎️|📞)\s*[:\s]*(\+?[\d•\*xX⁕\s-]{7,})`),
	regexp.MustCompile(`(\b[\d]*[\*xX•⁕]+[\d]{3,}\b|\b\d{10,}\b)`),
}

var servicePattern = regexp.MustCompile(`(?i)(?:Service|🔥 Service|Code)\s*(?:WhatsApp|Telegram|Google|Facebook|:|)\s*(\w+)`)
var nonDigit = regexp.MustCompile(`[^\d\*•xX⁕]`)
var spaceOrDash = regexp.MustCompile(`[\s-]`)
var onlyDigits = regexp.MustCompile(`\D`)

// otpWorker runs as a background goroutine
func (b *Bot) otpWorker() {
	logger.L.Info("OTP Worker started")

	// 1. Initial Login
	if b.scraper != nil {
		if err := b.scraper.Login(); err != nil {
			logger.L.Error("scraper login failed", "err", err)
		} else {
			logger.L.Info("scraper login successful")
		}
	}

	ticker := time.NewTicker(16 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-b.otpChan:
			b.processOTPMessage(msg.text, msg.chatID, msg.messageID, msg.replyMarkup)
		case <-ticker.C:
			if b.scraper != nil {
				b.pollScraper()
			}
		}
	}
}

func (b *Bot) pollScraper() {
	results, err := b.scraper.FetchSMS()
	if err != nil {
		logger.L.Error("scraper fetch failed", "err", err)
		// Try to re-login if session expired?
		_ = b.scraper.Login()
		return
	}

	if len(results) == 0 {
		logger.L.Debug("scraper found no messages")
		return
	}

	newCount := 0
	for _, res := range results {
		seen, err := b.processedSvc.IsSeen(res.Number)
		if err != nil {
			logger.L.Error("failed to check if number is seen", "err", err)
			continue
		}
		if seen {
			continue
		}

		// New SMS found!
		newCount++
		logger.L.Info("new SMS detected from scraper", "num", res.Number, "msg", res.Message)
		b.processScrapedSMS(res)
	}

	if newCount > 0 {
		logger.L.Info("scraper processing complete", "new_processed", newCount)
	}
}

func (b *Bot) processScrapedSMS(res SMSResult) {
	// 1. Extract details using ported logic
	otp := ExtractOTPCode(res.Message)
	shortCode, flag := DetectCountry(res.Number)
	masked := MaskPhoneNumber(res.Number)
	service := DetectServiceFromMessage(res.Message)
	if service == "UNKNOWN" && res.Service != "" && res.Service != "0" {
		service = strings.ToUpper(strings.ReplaceAll(res.Service, "_", " "))
	}
	icon := GetServiceAnimation(service)

	logger.L.Info("extracted SMS details", "num", res.Number, "otp", otp, "service", service, "country", shortCode)

	// 2. Mark as seen first to avoid duplicates
	err := b.processedSvc.Add(processednumber.ProcessedNumber{
		PhoneNumber: res.Number,
		OTPCode:     otp,
		ServiceName: service,
		Posted:      true,
	})
	if err != nil {
		logger.L.Error("failed to mark number as processed", "num", res.Number, "err", err)
	}

	// 3. Match with active number and notify user
	b.matchAndNotify(res.Number, otp, service)

	// 4. Forward to owners (for testing as requested)
	logger.L.Info("forwarding SMS to owners", "num", res.Number, "owners_count", len(b.ownerIDs))
	b.forwardToOwners(shortCode, flag, service, icon, masked, otp)
}

func (b *Bot) forwardToOwners(shortCode, flag, service, icon, masked, otp string) {
	msgText := fmt.Sprintf("%s #%s %s <code>%s</code>\n\n⛩️ 𝙿𝙾𝚆𝙴𝚁𝙴𝙳 𝙱𝚈 <a href=\"https://t.me/tamim_amv\">𝙏𝘼میم_amv</a> 👁",
		flag, shortCode, icon, masked)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔑 "+otp, "copy_otp::"+otp),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🤖 Number Bot", "https://t.me/sharknumber2bot"),
			tgbotapi.NewInlineKeyboardButtonURL("📺 Method", "https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz"),
		),
	)

	for _, ownerIDStr := range b.ownerIDs {
		var ownerChatID int64
		fmt.Sscanf(ownerIDStr, "%d", &ownerChatID)
		if ownerChatID != 0 {
			msg := tgbotapi.NewMessage(ownerChatID, msgText)
			msg.ParseMode = tgbotapi.ModeHTML
			msg.ReplyMarkup = keyboard
			msg.DisableWebPagePreview = true
			_, err := b.api.Send(msg)
			if err != nil {
				logger.L.Error("failed to forward SMS to owner", "owner", ownerIDStr, "err", err)
			} else {
				logger.L.Info("successfully forwarded SMS to owner", "owner", ownerIDStr)
			}
		}
	}
}

func (b *Bot) matchAndNotify(fullNumber, otp, service string) {
	// Matching logic similar to lines 147-177 in original otp_worker.go
	// But we have the full number now! (usually)
	// If res.Number is full, we can match exactly.

	allActive, err := b.activeSvc.GetAll()
	if err != nil {
		return
	}

	var matched *activenumber.ActiveNumber
	cleanFull := onlyDigits.ReplaceAllString(fullNumber, "")

	for _, an := range allActive {
		cleanActive := onlyDigits.ReplaceAllString(an.Number, "")
		if cleanActive == cleanFull || strings.HasSuffix(cleanFull, cleanActive) || strings.HasSuffix(cleanActive, cleanFull) {
			cp := an
			matched = &cp
			break
		}
	}

	if matched == nil {
		return
	}

	// Follow same steps as processOTPMessage (steps 179-275)
	// I'll refactor this into a common method in a follow-up if needed,
	// but for now let's keep it simple.

	foundUserID := matched.UserID
	menuMessageID := matched.MessageID
	platformFound := matched.Platform
	countryFound := matched.Country

	// Pre-assign next number, delete old, etc. (skipping for initial test if user only wants redirect to owner)
	// Actually, I'll keep the logic to ensure the user gets their OTP.

	var userChatID int64
	fmt.Sscanf(foundUserID, "%d", &userChatID)
	if userChatID != 0 {
		otpMsg := tgbotapi.NewMessage(userChatID,
			fmt.Sprintf("<b>✅ OTP Received for</b> <code>%s</code>\n\n<b>🔑 Your %s Code:</b> <code>%s</code>",
				fullNumber, service, otp))
		otpMsg.ParseMode = tgbotapi.ModeHTML
		_, _ = b.api.Send(otpMsg)
	}

	// Cleanup and next number assignment...
	_ = b.activeSvc.DeleteByNumber(matched.Number)
	nextNumber, _ := b.numberSvc.GetNextNumber(platformFound, countryFound, matched.Number)
	if nextNumber != "" {
		nextAN := activenumber.ActiveNumber{
			Number:    nextNumber,
			UserID:    foundUserID,
			Timestamp: time.Now(),
			MessageID: menuMessageID,
			Platform:  platformFound,
			Country:   countryFound,
		}
		_ = b.activeSvc.Insert(nextAN)
	}
}

func (b *Bot) processOTPMessage(text string, chatID int64, messageID int, markup *tgbotapi.InlineKeyboardMarkup) {
	// 1. Extract OTP code FIRST (for monitor and matching)
	otpCode := ""
	if markup != nil {
		for _, row := range markup.InlineKeyboard {
			for _, btn := range row {
				cleanBtn := onlyDigits.ReplaceAllString(btn.Text, "")
				if len(cleanBtn) >= 4 && len(cleanBtn) <= 8 {
					otpCode = cleanBtn
					break
				}
			}
			if otpCode != "" {
				break
			}
		}
	}
	if otpCode == "" {
		for _, p := range otpPatterns {
			m := p.FindStringSubmatch(text)
			if len(m) > 1 {
				code := spaceOrDash.ReplaceAllString(m[1], "")
				if len(code) >= 4 {
					otpCode = code
					break
				}
			}
		}
	}

	// 2. Extract masked phone number
	detectedNum := ""
	for _, p := range numberPatterns {
		m := p.FindStringSubmatch(text)
		if len(m) > 1 {
			cleaned := strings.ReplaceAll(m[1], " ", "")
			cleaned = strings.ReplaceAll(cleaned, "-", "")
			digitOnly := onlyDigits.ReplaceAllString(cleaned, "")
			hasMask := strings.ContainsAny(cleaned, "•*xX⁕")
			if hasMask || len(digitOnly) >= 7 {
				detectedNum = cleaned
				break
			}
		}
	}

	// 3. Forward to owners if it's the test group
	const testGroupID = -1003422191454
	const dummyTestGroupID = -1003678266458
	if chatID == testGroupID || chatID == dummyTestGroupID {
		logger.L.Info("Forwarding message from test group to owners", "chat", chatID, "owner_count", len(b.ownerIDs))
		monitorMsg := fmt.Sprintf("<b>🔍 Test Group Monitor</b>\n\n%s", text)
		if otpCode != "" {
			monitorMsg += fmt.Sprintf("\n\n<b>🔑 Extracted OTP:</b> <code>%s</code>", otpCode)
		} else {
			monitorMsg += "\n\n<b>❌ No OTP detected in this message.</b>"
		}

		for _, ownerIDStr := range b.ownerIDs {
			var ownerChatID int64
			fmt.Sscanf(ownerIDStr, "%d", &ownerChatID)
			if ownerChatID != 0 {
				markupRow1 := []tgbotapi.InlineKeyboardButton{}
				if otpCode != "" {
					markupRow1 = append(markupRow1, tgbotapi.NewInlineKeyboardButtonData("🔑 "+otpCode, "copy_otp::"+otpCode))
				} else {
					markupRow1 = append(markupRow1, tgbotapi.NewInlineKeyboardButtonData("❌ No OTP", "no_otp"))
				}

				markup := tgbotapi.NewInlineKeyboardMarkup(
					markupRow1,
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL("🤖 Number Bot", "https://t.me/sharknumber2bot"),
						tgbotapi.NewInlineKeyboardButtonURL("📺 Method", "https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz"),
					),
				)

				msg := tgbotapi.NewMessage(ownerChatID, monitorMsg)
				msg.ParseMode = tgbotapi.ModeHTML
				msg.ReplyMarkup = markup

				_, err := b.api.Send(msg)
				if err != nil {
					// Fallback: If markup still fails, send without it
					msg.ReplyMarkup = nil
					_, _ = b.api.Send(msg)
					logger.L.Error("failed to forward msg to owner with markup, sent without", "owner", ownerIDStr, "err", err)
				} else {
					logger.L.Info("successfully forwarded msg to owner with design", "owner", ownerIDStr)
				}
			}
		}
	}

	if otpCode == "" || detectedNum == "" {
		return
	}

	logger.L.Info("OTP Worker parsed", "code", otpCode, "num", detectedNum)

	// 3. Build prefix/suffix for matching
	cleaned := nonDigit.ReplaceAllString(detectedNum, "")
	parts := regexp.MustCompile(`[*•xX⁕]+`).Split(cleaned, -1)
	prefix := ""
	suffix := ""
	if len(parts) > 0 {
		prefix = parts[0]
	}
	if len(parts) > 1 {
		suffix = parts[len(parts)-1]
	}

	// 4. Find matching active number
	allActive, err := b.activeSvc.GetAll()
	if err != nil {
		logger.L.Error("OTP Worker GetAll failed", "err", err)
		return
	}

	var matched *activenumber.ActiveNumber
	for _, an := range allActive {
		cleanActive := onlyDigits.ReplaceAllString(an.Number, "")
		if strings.HasPrefix(cleanActive, prefix) && strings.HasSuffix(cleanActive, suffix) {
			cp := an // copy
			matched = &cp
			break
		}
	}

	if matched == nil {
		return
	}

	// ✅ FIX: Capture user/message info BEFORE any deletion
	foundUserID := matched.UserID
	menuMessageID := matched.MessageID
	platformFound := matched.Platform
	countryFound := matched.Country
	fullNumber := matched.Number

	logger.L.Info("OTP matched", "number", fullNumber, "user", foundUserID)

	// 5. Pre-assign next number (before deleting old)
	nextNumber, _ := b.numberSvc.GetNextNumber(platformFound, countryFound, fullNumber)

	// 6. Delete old number from active (release it)
	_ = b.activeSvc.DeleteByNumber(fullNumber)

	// 7. Insert pre-assigned next number if available
	if nextNumber != "" {
		nextAN := activenumber.ActiveNumber{
			Number:    nextNumber,
			UserID:    foundUserID,
			Timestamp: time.Now(),
			MessageID: menuMessageID,
			Platform:  platformFound,
			Country:   countryFound,
		}
		_ = b.activeSvc.Insert(nextAN)
		_ = b.seenSvc.Add(foundUserID, nextNumber, countryFound)
	}

	// 8. Update stats
	_ = b.statsSvc.IncrOtpStat(countryFound)
	_ = b.statsSvc.IncrUserOtpStat(foundUserID, countryFound)

	// 9. Parse userID as int64 to send messages
	var userChatID int64
	fmt.Sscanf(foundUserID, "%d", &userChatID)
	if userChatID == 0 {
		return
	}

	// 10. ✅ FIX: Send OTP to user's private inbox — CORRECT bot.Send call
	serviceMatch := servicePattern.FindStringSubmatch(text)
	displayService := platformFound
	if len(serviceMatch) > 1 {
		displayService = serviceMatch[1]
	}
	if displayService == "" {
		displayService = "OTP"
	}

	otpMsg := tgbotapi.NewMessage(userChatID,
		fmt.Sprintf("<b>✅ OTP Received for</b> <code>%s</code>\n\n<b>🔑 Your %s Code:</b> <code>%s</code>",
			fullNumber, displayService, otpCode))
	otpMsg.ParseMode = tgbotapi.ModeHTML
	if _, err := b.api.Send(otpMsg); err != nil {
		logger.L.Error("failed to send OTP to user", "user", foundUserID, "err", err)
	} else {
		logger.L.Info("OTP sent to user", "user", foundUserID)
	}

	// 11. Update the menu message to show new number state
	if menuMessageID != 0 {
		// Get current numbers for user
		currentActives, _ := b.activeSvc.GetByUser(foundUserID)
		numDisplay := ""
		for _, a := range currentActives {
			numDisplay += fmt.Sprintf("<code>%s</code>\n", a.Number)
		}

		menuText := fmt.Sprintf("<b>%s (%s) Number(s) Assigned:</b>\n%s\n<b>Waiting for OTP...</b>",
			countryFound, platformFound, numDisplay)

		groupLink := b.settingsSvc.GetGroupLink()
		menuMarkup := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔄 Change Number",
					fmt.Sprintf("change_number::%s::%s", platformFound, countryFound)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("OTP Groupe 👥", groupLink),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⬅️ Back",
					fmt.Sprintf("back_to_countries::%s", platformFound)),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(userChatID, int(menuMessageID), menuText)
		editMsg.ParseMode = tgbotapi.ModeHTML
		editMsg.ReplyMarkup = &menuMarkup
		if _, err := b.api.Send(editMsg); err != nil {
			errStr := err.Error()
			if !strings.Contains(errStr, "message is not modified") {
				logger.L.Error("menu edit failed", "err", err)
			}
		}
	}

	// suppress unused import warning
	_ = sql.ErrNoRows
}

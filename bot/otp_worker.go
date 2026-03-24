package bot

import (
	"fmt"
	"regexp"
	"shark_bot/pkg/logger"
	"strings"
	"time"

	"math/rand"
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

// Keep owner forwarding code available, but disabled for now.
const forwardOTPToOwnersEnabled = false

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

	for {
		if b.scraper != nil {
			b.pollScraper()
		}
		// 16s base + 0-4s jitter
		jitter := time.Duration(rand.Intn(4000)) * time.Millisecond
		wait := 16*time.Second + jitter
		log.Info("scraper waiting", "duration", wait.String())
		time.Sleep(wait)
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

	// 4. Forward to owners (temporarily disabled)
	if forwardOTPToOwnersEnabled {
		logger.L.Info("forwarding SMS to owners", "num", res.Number, "owners_count", len(b.ownerIDs))
		b.forwardToOwners(shortCode, flag, service, icon, masked, otp)
	} else {
		logger.L.Info("owner OTP forwarding disabled", "num", res.Number)
	}
}

func (b *Bot) forwardToOwners(shortCode, flag, service, icon, masked, otp string) {
	msgText := fmt.Sprintf("%s #%s %s <code>%s</code>\n\n⛩️ 𝙿𝙾𝚆𝙴𝚁𝙴𝙳 𝙱𝚈 <a href=\"https://t.me/zvoidois\">𝒵𝒶𝒽𝒾𝒹</a> 👁",
		flag, shortCode, icon, masked)

	// Custom button types to support copy_text which is missing in the library
	type CopyTextButton struct {
		Text string `json:"text"`
	}
	type CustomButton struct {
		Text         string          `json:"text"`
		CallbackData string          `json:"callback_data,omitempty"`
		URL          string          `json:"url,omitempty"`
		CopyText     *CopyTextButton `json:"copy_text,omitempty"`
	}

	keyboard := struct {
		InlineKeyboard [][]CustomButton `json:"inline_keyboard"`
	}{
		InlineKeyboard: [][]CustomButton{
			{
				{Text: otp, CopyText: &CopyTextButton{Text: otp}},
			},
			{
				{Text: "🤖 Number Bot", URL: "https://t.me/sharknumber2bot"},
				{Text: "📺 Method", URL: "https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz"},
			},
		},
	}

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
		// Custom button types to support copy_text
		type CopyTextButton struct {
			Text string `json:"text"`
		}
		type CustomButton struct {
			Text     string          `json:"text"`
			CopyText *CopyTextButton `json:"copy_text,omitempty"`
		}

		otpMsg.ReplyMarkup = struct {
			InlineKeyboard [][]CustomButton `json:"inline_keyboard"`
		}{
			InlineKeyboard: [][]CustomButton{
				{
					{Text: otp, CopyText: &CopyTextButton{Text: otp}},
				},
			},
		}
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

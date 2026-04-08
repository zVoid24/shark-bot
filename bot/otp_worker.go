package bot

import (
	"fmt"
	"regexp"
	"shark_bot/pkg/logger"
	"strings"
	"time"

	"math/rand"
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
	logger.L.Info("OTP Worker started", "scraper_count", len(b.scrapers))

	for _, s := range b.scrapers {
		go b.runScraperLoop(s)
	}
}

func (b *Bot) runScraperLoop(s *Scraper) {
	logger.L.Info("starting scraper loop", "user", s.username)

	// 1. Initial Login
	if err := s.Login(); err != nil {
		logger.L.Error("scraper login failed", "user", s.username, "err", err)
	} else {
		logger.L.Info("scraper login successful", "user", s.username)
	}

	for {
		b.pollScraper(s)
		// 20s base + 0-5s jitter to avoid website blocking
		jitter := time.Duration(rand.Intn(5000)) * time.Millisecond
		wait := 20*time.Second + jitter
		logger.L.Debug("scraper waiting", "user", s.username, "duration", wait.String())
		time.Sleep(wait)
	}
}

func (b *Bot) pollScraper(s *Scraper) {
	results, err := s.FetchSMS()
	if err != nil {
		logger.L.Error("scraper fetch failed", "user", s.username, "err", err)
		// Try to re-login if session expired?
		_ = s.Login()
		return
	}

	if len(results) == 0 {
		logger.L.Debug("scraper found no messages")
		return
	}

	newCount := 0
	oldSkipped := 0

	// 1. Find the newest message as the reference "Now" for the scraper
	var latestTime time.Time
	const timeLayout = "2006-01-02 15:04:05"
	for _, res := range results {
		t, err := time.ParseInLocation(timeLayout, res.DateTime, time.Local)
		if err == nil {
			if t.After(latestTime) {
				latestTime = t
			}
		}
	}

	// 2. Log timing info for debugging
	if !latestTime.IsZero() {
		offset := time.Since(latestTime)
		logger.L.Debug("scraper timing info",
			"user", s.username,
			"latest_sms", latestTime.Format(timeLayout),
			"server_now", time.Now().Format(timeLayout),
			"detected_offset", offset.String(),
		)
	}

	for _, res := range results {
		// 3. Skip if older than 15 minutes relative to the NEWEST message found
		if !latestTime.IsZero() {
			msgTime, err := time.ParseInLocation(timeLayout, res.DateTime, time.Local)
			if err == nil {
				if latestTime.Sub(msgTime) > 15*time.Minute {
					oldSkipped++
					continue
				}
			}
		}

		otp := ExtractOTPCode(res.Message)
		seen, err := b.processedSvc.IsSeen(res.Number, otp)
		if err != nil {
			logger.L.Error("failed to check if number is seen", "err", err)
			continue
		}
		if seen {
			continue
		}

		// New SMS found!
		newCount++
		logger.L.Info("new SMS detected from scraper", "num", res.Number, "msg", res.Message, "user", s.username)
		b.processScrapedSMS(res)
	}

	logger.L.Info("scraper poll complete",
		"user", s.username,
		"total", len(results),
		"new", newCount,
		"skipped_old", oldSkipped,
		"skipped_dup", len(results)-newCount-oldSkipped,
	)
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

	logger.L.Info("extracted SMS details", "num", res.Number, "otp", otp, "service", service, "country", shortCode, "user", res.Account)

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

	// 3. Send directly to central group chat
	b.sendToCentralGroup(shortCode, flag, service, icon, masked, otp, res.Account)

	// 4. Add a small delay between messages to respect Telegram rate limits
	time.Sleep(2 * time.Second)
}

func (b *Bot) sendToCentralGroup(shortCode, flag, service, icon, masked, otp, user string) {
	msgText := fmt.Sprintf("%s #%s %s <code>%s</code>\n\n<tg-emoji emoji-id='5888699182734122090'>⛩️</tg-emoji> 𝙿𝙾𝚆𝙴𝚁𝙴𝙳 𝙱𝚈 <a href=\"https://t.me/shark_sms\">𝙍𝙄𝙕𝙑𝙄</a> <tg-emoji emoji-id='5888704237910627502'>👁</tg-emoji>",
		flag, shortCode, icon, masked)

	// Custom button types to support copy_text and custom_emoji_id
	type CopyTextButton struct {
		Text string `json:"text"`
	}
	type CustomButton struct {
		Text          string          `json:"text"`
		URL           string          `json:"url,omitempty"`
		CopyText      *CopyTextButton `json:"copy_text,omitempty"`
		CustomEmojiID string          `json:"icon_custom_emoji_id,omitempty"`
	}

	keyboard := struct {
		InlineKeyboard [][]CustomButton `json:"inline_keyboard"`
	}{
		InlineKeyboard: [][]CustomButton{
			{
				{
					Text:          otp,
					CopyText:      &CopyTextButton{Text: otp},
					CustomEmojiID: "6176966310920983412",
				},
			},
			{
				{
					Text:          "Number Bot",
					URL:           "https://t.me/sharknumber2bot",
					CustomEmojiID: "5231197925178089666",
				},
				{
					Text:          "Method",
					URL:           "https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz",
					CustomEmojiID: "5942902988564600402",
				},
			},
		},
	}

	msg := tgbotapi.NewMessage(b.otpTargetChatID, msgText)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	msg.DisableWebPagePreview = true

	// Retry loop for rate limits
	for retries := 0; retries < 3; retries++ {
		_, err := b.api.Send(msg)
		if err != nil {
			if apiErr, ok := err.(*tgbotapi.Error); ok && apiErr.Code == 429 {
				waitSecs := apiErr.RetryAfter
				if waitSecs <= 0 {
					waitSecs = 5
				}
				logger.L.Warn("rate limited by telegram, retrying", "wait_secs", waitSecs, "retry", retries+1)
				time.Sleep(time.Duration(waitSecs) * time.Second)
				continue
			}
			logger.L.Error("failed to send OTP to central group", "chat_id", b.otpTargetChatID, "err", err, "user", user)
			break
		}
		logger.L.Info("successfully sent OTP to central group", "chat_id", b.otpTargetChatID, "user", user)
		break
	}
}

// matchAndNotify is disabled in OTP-only mode.
func (b *Bot) matchAndNotify(fullNumber, otp, service string) {
	// Logic disabled
}

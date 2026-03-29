package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"shark_bot/bot"
	"shark_bot/config"
	"shark_bot/infra/db"
	"shark_bot/infra/repository"
	"shark_bot/internal/activenumber"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

const defaultUserID = "5160630771"

var digitOnly = regexp.MustCompile(`\D`)

func normalizeNumber(v string) string {
	return digitOnly.ReplaceAllString(v, "")
}

func main() {
	userID := flag.String("user", defaultUserID, "Target Telegram user ID")
	number := flag.String("number", "", "Optional incoming scraper number. If empty, first active number of user is used")
	otp := flag.String("otp", "123456", "OTP code to send")
	service := flag.String("service", "WhatsApp", "Service name in OTP message")
	dryRun := flag.Bool("dry-run", false, "If true, do not send Telegram message")
	flag.Parse()

	cnf := config.Load()
	ctx := context.Background()

	dbConn, err := db.NewConnection(&cnf.Database)
	if err != nil {
		panic(fmt.Errorf("db connection failed: %w", err))
	}
	defer dbConn.Close()

	activeRepo := repository.NewActiveNumberRepo(dbConn)
	activeSvc := activenumber.NewService(activeRepo)

	var cache *bot.ActiveNumberCache
	redisClient := redis.NewClient(&redis.Options{
		Addr:      cnf.Redis.Addr,
		Password:  cnf.Redis.Password,
		DB:        cnf.Redis.DB,
		TLSConfig: tlsConfig(cnf.Redis.EnableTLS),
	})
	if err := redisClient.Ping(ctx).Err(); err == nil {
		cache = bot.NewActiveNumberCache(
			redisClient,
			cnf.Redis.KeyPrefix,
			time.Duration(cnf.Redis.ActiveTTL)*time.Second,
		)
		fmt.Printf("Redis connected: %s\n", cnf.Redis.Addr)
	} else {
		fmt.Printf("Redis unavailable (%v). Falling back to DB only.\n", err)
	}

	targetActive, err := pickActiveNumber(activeSvc, *userID, *number)
	if err != nil {
		panic(fmt.Errorf("failed to choose active number for simulation: %w", err))
	}

	incomingNumber := targetActive.Number
	if *number != "" {
		incomingNumber = *number
	}

	fmt.Printf("Simulating scraper OTP: number=%s otp=%s service=%s\n", incomingNumber, *otp, *service)

	matched, err := resolveMatch(ctx, cache, activeSvc, incomingNumber)
	if err != nil {
		panic(err)
	}
	if matched == nil {
		fmt.Printf("No match for provided number %s. Falling back to user's active number %s\n", incomingNumber, targetActive.Number)
		incomingNumber = targetActive.Number
		matched, err = resolveMatch(ctx, cache, activeSvc, incomingNumber)
		if err != nil {
			panic(err)
		}
		if matched == nil {
			panic("no active-number match found even after fallback to user's active number")
		}
	}

	chatID, err := strconv.ParseInt(matched.UserID, 10, 64)
	if err != nil {
		panic(fmt.Errorf("invalid matched user id %q: %w", matched.UserID, err))
	}

	text := fmt.Sprintf("<b>✅ OTP Received for</b> <code>%s</code>\n\n<b>🔑 Your %s Code:</b> <code>%s</code>",
		incomingNumber, *service, *otp)

	if *dryRun {
		fmt.Printf("[DRY RUN] Would send to user %s (chat_id=%d):\n%s\n", matched.UserID, chatID, text)
		return
	}

	api, err := tgbotapi.NewBotAPI(cnf.Telegram.BotToken)
	if err != nil {
		panic(fmt.Errorf("telegram bot init failed: %w", err))
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML

	if _, err := api.Send(msg); err != nil {
		panic(fmt.Errorf("failed to send simulated otp message: %w", err))
	}

	fmt.Printf("OTP message sent to user %s for number %s\n", matched.UserID, incomingNumber)
}

func tlsConfig(enabled bool) *tls.Config {
	if !enabled {
		return nil
	}
	return &tls.Config{MinVersion: tls.VersionTLS12}
}

func pickActiveNumber(activeSvc *activenumber.Service, userID, incomingNumber string) (*activenumber.ActiveNumber, error) {
	acts, err := activeSvc.GetByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to read active numbers for user %s: %w", userID, err)
	}
	if len(acts) == 0 {
		return nil, fmt.Errorf("no active number found for user %s", userID)
	}
	userPrimary := acts[0]

	if incomingNumber == "" {
		return &userPrimary, nil
	}

	all, err := activeSvc.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read all active numbers: %w", err)
	}
	in := normalizeNumber(incomingNumber)
	for _, an := range all {
		n := normalizeNumber(an.Number)
		if n == in || (in != "" && (hasSuffix(in, n) || hasSuffix(n, in))) {
			cp := an
			return &cp, nil
		}
	}

	fmt.Printf("Warning: provided number %s not found in active_numbers. Using user's active number %s.\n", incomingNumber, userPrimary.Number)
	return &userPrimary, nil
}

func resolveMatch(ctx context.Context, cache *bot.ActiveNumberCache, activeSvc *activenumber.Service, incomingNumber string) (*activenumber.ActiveNumber, error) {
	if cache != nil {
		matched, err := cache.GetByNumber(ctx, incomingNumber)
		if err == nil && matched != nil {
			fmt.Println("Matched via Redis cache")
			return matched, nil
		}
		if err != nil {
			fmt.Printf("Redis lookup failed (%v). Falling back to DB.\n", err)
		}
	}

	all, err := activeSvc.GetAll()
	if err != nil {
		return nil, err
	}

	in := normalizeNumber(incomingNumber)
	for _, an := range all {
		n := normalizeNumber(an.Number)
		if n == in || (in != "" && (hasSuffix(in, n) || hasSuffix(n, in))) {
			cp := an
			if cache != nil {
				_ = cache.Set(ctx, cp)
			}
			fmt.Println("Matched via DB fallback")
			return &cp, nil
		}
	}
	return nil, nil
}

func hasSuffix(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if len(b) > len(a) {
		return false
	}
	return a[len(a)-len(b):] == b
}


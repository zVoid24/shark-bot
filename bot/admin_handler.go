package bot

import (
	"context"
	"fmt"
	"shark_bot/pkg/logger"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const adminHelp = `<b>--- Admin Command Panel ---</b>

<b><u>Admin Management</u></b>
<code>/addadmin [user_id]</code> - Add a user to admin list.
<code>/rmvadmin [user_id]</code> - Remove a user from admin list.
<code>/adminlist</code> - Show all current admins.

<b><u>Number Management</u></b>
<code>/addnumber</code> - Add numbers to a platform and country.
<code>/removenumber</code> - Remove a country's numbers from a platform.
<code>/numberlimit [Plat] [Coun] [Limit]</code> - Set limit per user.
<code>/numberremove [Plat] [Coun] [on/off]</code> - Set delete policy on change.

<b><u>Configuration</u></b>
<code>/cancel</code> - Cancel the current operation.

<b><u>User Management</u></b>
<code>/block [user_id]</code> - Block a user.
<code>/unblock [user_id]</code> - Unblock a user.
<code>/unblockall</code> - Unblock all users.
<code>/blocklist</code> - Show all blocked users.
<code>/seestatus [user_id]</code> - See a user's OTP stats.

<b><u>Broadcast &amp; Stats</u></b>
<code>/all [message]</code> - Broadcast to all users.
<code>/statusall</code> - Show total OTP stats.

<b><u>Bot Management</u></b>
<code>/resetall</code> - Reset all user data.
<code>/admin</code> or <code>/cmd</code> - Show this help.`

// isAdmin checks if user is an admin
func (b *Bot) isAdmin(userID string) bool {
	ok, err := b.adminSvc.IsAdmin(userID)
	if err != nil {
		logger.L.Error("isAdmin check failed", "err", err)
	}
	return ok
}

func (b *Bot) handleAdminCmd(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	b.sendHTML(msg.Chat.ID, adminHelp)
}

func (b *Bot) handleAdminList(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	admins, _ := b.adminSvc.GetAll()
	lines := "<b>--- Current Admin List ---</b>\n\n"
	for _, uid := range admins {
		lines += fmt.Sprintf("• <code>%s</code>\n", uid)
	}
	b.sendHTML(msg.Chat.ID, lines)
}

func (b *Bot) handleAddAdmin(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	args := strings.Fields(msg.CommandArguments())
	if len(args) == 0 {
		b.sendHTML(msg.Chat.ID, "<b>Usage: /addadmin [user_id]</b>")
		return
	}
	uid := args[0]
	if err := b.adminSvc.Add(uid); err != nil {
		b.sendHTML(msg.Chat.ID, "<b>Error adding admin.</b>")
		return
	}
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>User <code>%s</code> has been added as an admin.</b>", uid))
}

func (b *Bot) handleRemoveAdmin(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	args := strings.Fields(msg.CommandArguments())
	if len(args) == 0 {
		b.sendHTML(msg.Chat.ID, "<b>Usage: /rmvadmin [user_id]</b>")
		return
	}
	uid := args[0]
	for _, ownerID := range b.ownerIDs {
		if ownerID == uid {
			b.sendHTML(msg.Chat.ID, "<b>Error: Principal Owner cannot be removed.</b>")
			return
		}
	}
	count, _ := b.adminSvc.Count()
	if count <= 1 {
		b.sendHTML(msg.Chat.ID, "<b>Cannot remove the last admin.</b>")
		return
	}
	_ = b.adminSvc.Remove(uid)
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>User <code>%s</code> removed from admins.</b>", uid))
}

func (b *Bot) handleBlock(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	args := strings.Fields(msg.CommandArguments())
	if len(args) == 0 {
		b.sendHTML(msg.Chat.ID, "<b>Usage: /block [user_id]</b>")
		return
	}
	uid := args[0]
	_ = b.userSvc.BlockUser(uid)
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>User <code>%s</code> has been blocked.</b>", uid))
}

func (b *Bot) handleUnblock(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	args := strings.Fields(msg.CommandArguments())
	if len(args) == 0 {
		b.sendHTML(msg.Chat.ID, "<b>Usage: /unblock [user_id]</b>")
		return
	}
	uid := args[0]
	_ = b.userSvc.UnblockUser(uid)
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>User <code>%s</code> has been unblocked.</b>", uid))
}

func (b *Bot) handleUnblockAll(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	_ = b.userSvc.UnblockAll()
	b.sendHTML(msg.Chat.ID, "<b>Successfully unblocked all users.</b>")
}

func (b *Bot) handleBlockList(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	blocked, _ := b.userSvc.GetBlockedUsers()
	if len(blocked) == 0 {
		b.sendHTML(msg.Chat.ID, "<b>There are no blocked users.</b>")
		return
	}
	text := "<b>--- Blocked Users List ---</b>\n\n"
	for _, uid := range blocked {
		text += fmt.Sprintf("<b>ID:</b> <code>%s</code>\n", uid)
	}
	b.sendHTML(msg.Chat.ID, text)
}

func (b *Bot) handleSeeStatus(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	args := strings.Fields(msg.CommandArguments())
	if len(args) == 0 {
		b.sendHTML(msg.Chat.ID, "<b>Usage: /seestatus [user_id]</b>")
		return
	}
	uid := args[0]
	user, _ := b.userSvc.GetUser(uid)
	statusStr := "🟢 Active"
	if user != nil && user.IsBlocked {
		statusStr = "🔴 Blocked"
	}
	actives, _ := b.activeSvc.GetByUser(uid)
	holdingNum := "None"
	if len(actives) > 0 {
		holdingNum = actives[0].Number
	}
	text := fmt.Sprintf("<b>Status for User ID:</b> <code>%s</code>\n\n<b>Status:</b> %s\n<b>Holding:</b> <code>%s</code>\n\n<b>--- User OTP Stats ---</b>\n",
		uid, statusStr, holdingNum)
	stats, _ := b.statsSvc.GetUserOtpStats(uid)
	if len(stats) == 0 {
		text += "<b>No OTPs received by this user yet.</b>"
	} else {
		for _, s := range stats {
			text += fmt.Sprintf("<b>%s:</b> <code>%d</code> <b>OTPs</b>\n", s.Country, s.Count)
		}
	}
	b.sendHTML(msg.Chat.ID, text)
}

func (b *Bot) handleStatusAll(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	stats, _ := b.statsSvc.GetAllOtpStats()
	text := "<b>📊 Total OTP Status 📊</b>\n\n"
	total := 0
	for _, s := range stats {
		text += fmt.Sprintf("<b>%s:</b> <code>%d</code> <b>OTPs</b>\n", s.Country, s.Count)
		total += s.Count
	}
	text += fmt.Sprintf("\n<b>Total All OTPs:</b> <code>%d</code>", total)
	b.sendHTML(msg.Chat.ID, text)
}

func (b *Bot) handleResetAll(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	if b.activeCache != nil {
		acts, _ := b.activeSvc.GetAll()
		ctx := context.Background()
		for _, an := range acts {
			_ = b.activeCache.DeleteByNumber(ctx, an.Number)
		}
	}
	_ = b.activeSvc.DeleteAll()
	_ = b.statsSvc.ResetAll()
	_ = b.seenSvc.ResetAll()
	b.sendHTML(msg.Chat.ID, "<b>All user data has been successfully reset.</b>")
}

func (b *Bot) handleBroadcast(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	text := msg.CommandArguments()
	if text == "" {
		b.sendHTML(msg.Chat.ID, "<b>Usage: /all [your message]</b>")
		return
	}
	dbUserIDs, _ := b.userSvc.GetUnblockedUserIDs()
	registryUserIDs, err := getKnownUserIDs()
	if err != nil {
		log.Warn("failed to read user registry", "err", err)
	}

	userIDs := mergeUniqueUserIDs(dbUserIDs, registryUserIDs)
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>Starting broadcast to %d users...</b>", len(userIDs)))
	go b.runBroadcast(userIDs, text, msg.Chat.ID)
}

func (b *Bot) runBroadcast(userIDs []string, text string, adminChatID int64) {
	var (
		successCount int
		blockedCount int
		errorCount   int
		mu           sync.Mutex
		wg           sync.WaitGroup
	)

	userChan := make(chan string, 100)
	workerCount := 30
	// Telegram allows 30 messages per second. We use ~29/sec to be safe.
	ticker := time.NewTicker(34 * time.Millisecond)
	defer ticker.Stop()

	// Start worker pool
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for uid := range userChan {
				// Throttle based on ticker
				<-ticker.C

				chatID, err := strconv.ParseInt(uid, 10, 64)
				if err != nil || chatID == 0 {
					continue
				}

				m := tgbotapi.NewMessage(chatID, text)
				m.ParseMode = tgbotapi.ModeHTML

				_, err = b.api.Send(m)
				mu.Lock()
				if err != nil {
					// Check if user blocked the bot
					errStr := strings.ToLower(err.Error())
					if strings.Contains(errStr, "forbidden") || strings.Contains(errStr, "blocked") || strings.Contains(errStr, "deactivated") {
						blockedCount++
						// Persist the block status in our DB to skip them in future broadcasts
						_ = b.userSvc.BlockUser(uid)
					} else {
						errorCount++
						log.Error("broadcast send failed", "user_id", uid, "err", err)
					}
				} else {
					successCount++
				}
				mu.Unlock()
			}
		}()
	}

	// Feed workers
	for _, uid := range userIDs {
		userChan <- uid
	}
	close(userChan)

	// Wait for all messages to be processed
	wg.Wait()

	// Send final summary to admin
	summary := fmt.Sprintf("<b>📢 Broadcast Summary</b>\n\n"+
		"✅ <b>Delivered:</b> <code>%d</code>\n"+
		"🚫 <b>Blocked/Deactivated:</b> <code>%d</code>\n"+
		"❌ <b>Errors:</b> <code>%d</code>\n\n"+
		"👥 <b>Total Users Targetted:</b> <code>%d</code>",
		successCount, blockedCount, errorCount, len(userIDs))

	b.sendHTML(adminChatID, summary)
	log.Info("broadcast completed", "success", successCount, "blocked", blockedCount, "errors", errorCount)
}

func (b *Bot) handleSetNumberLimit(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	var plat, coun string
	var limit int
	n, _ := fmt.Sscanf(msg.CommandArguments(), "%s %s %d", &plat, &coun, &limit)
	if n < 3 || limit <= 0 {
		b.sendHTML(msg.Chat.ID, "<b>Usage: /numberlimit [Platform] [Country] [Limit]</b>")
		return
	}
	_ = b.settingsSvc.SetNumberLimit(plat, coun, limit)
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>✅ Limit Updated!</b>\n\n<b>Platform:</b> %s\n<b>Country:</b> %s\n<b>Max Numbers:</b> %d", plat, coun, limit))
}

func (b *Bot) handleToggleRemovePolicy(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	parts := strings.Fields(msg.CommandArguments())
	if len(parts) != 3 || (parts[2] != "on" && parts[2] != "off") {
		b.sendHTML(msg.Chat.ID, "<b>Usage: /numberremove [Platform] [Country] [on/off]</b>")
		return
	}
	plat, coun, status := parts[0], parts[1], parts[2]
	_ = b.settingsSvc.SetRemovePolicy(plat, coun, status)
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>✅ Configuration Updated!</b>\n\n<b>Platform:</b> %s\n<b>Country:</b> %s\n<b>Remove on Change:</b> %s", plat, coun, strings.ToUpper(status)))
}

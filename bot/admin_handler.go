package bot

import (
	"fmt"
	"shark_bot/pkg/logger"
	"strings"

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
<code>/cancel</code> - Cancel the current operation.

<b><u>Configuration</u></b>
<code>/addchatid</code> - Add new Target Group Chat IDs.
<code>/setgroupe [link]</code> - Set the OTP Group link.

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
	userIDs, _ := b.userSvc.GetAllUserIDs()
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>Starting broadcast to %d users...</b>", len(userIDs)))
	go b.runBroadcast(userIDs, text)
}

func (b *Bot) runBroadcast(userIDs []string, text string) {
	batchSize := 20
	for i := 0; i < len(userIDs); i += batchSize {
		end := i + batchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		for _, uid := range userIDs[i:end] {
			var chatID int64
			fmt.Sscanf(uid, "%d", &chatID)
			if chatID == 0 {
				continue
			}
			m := tgbotapi.NewMessage(chatID, text)
			m.ParseMode = tgbotapi.ModeHTML
			b.api.Send(m)
		}
	}
}

func (b *Bot) handleSetGroupLink(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	args := strings.Fields(msg.CommandArguments())
	if len(args) == 0 || !strings.HasPrefix(args[0], "http") {
		b.sendHTML(msg.Chat.ID, "<b>Usage: /setgroupe [link]</b>")
		return
	}
	_ = b.settingsSvc.Set("group_link", args[0])
	b.sendHTML(msg.Chat.ID, fmt.Sprintf("<b>OTP Group link updated to:</b>\n%s", args[0]))
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

func (b *Bot) handleAddChatID(msg *tgbotapi.Message) {
	if !b.isAdmin(fmt.Sprintf("%d", msg.From.ID)) {
		return
	}
	// Enter conversation state for adding chat IDs
	b.setConvState(msg.From.ID, &convContext{Step: convStepAddChatID})
	b.removeKeyboard(msg.Chat.ID, "<b>Enter the new Target Group Chat ID(s), one per line.</b>")
}

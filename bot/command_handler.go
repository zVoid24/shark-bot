package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCommand dispatches bot commands to the correct handler
func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.handleStart(msg)
	case "admin", "cmd":
		b.handleAdminCmd(msg)
	case "adminlist":
		b.handleAdminList(msg)
	case "addadmin":
		b.handleAddAdmin(msg)
	case "rmvadmin":
		b.handleRemoveAdmin(msg)
	case "block":
		b.handleBlock(msg)
	case "unblock":
		b.handleUnblock(msg)
	case "unblockall":
		b.handleUnblockAll(msg)
	case "blocklist":
		b.handleBlockList(msg)
	case "seestatus":
		b.handleSeeStatus(msg)
	case "statusall":
		b.handleStatusAll(msg)
	case "resetall":
		b.handleResetAll(msg)
	case "all":
		b.handleBroadcast(msg)
	case "numberlimit":
		b.handleSetNumberLimit(msg)
	case "numberremove":
		b.handleToggleRemovePolicy(msg)
	case "addnumber":
		b.handleAddNumber(msg)
	case "removenumber":
		b.handleRemoveNumber(msg)
	case "cancel":
		b.handleCancel(msg)
	case "setotpprice":
		b.handleSetOTPPrice(msg)
	case "setminwithdraw":
		b.handleSetMinWithdraw(msg)
	case "getnumber":
		b.handleGetNumber(msg)
	}
}

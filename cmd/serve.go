package main

import (
	"log"
	"shark_bot/config"
	"shark_bot/infra/db"
	utils "shark_bot/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Serve() {
	cnf := config.Load()
	db, err := db.NewConnection(&cnf.Database)
	if err != nil {
		log.Fatal("Unable to connect with database")
		panic(err)
	}
	bot, err := tgbotapi.NewBotAPI(cnf.Telegram.BotToken)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	utilHandler := utils.NewHandler(bot)
}

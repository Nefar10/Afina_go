package main

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var gBot *tgbotapi.BotAPI
var gToken string
var gChatid int64
var gOwner int64

func init() {

	var err error
	gToken := os.Getenv(TOKEN_NAME_IN_OS)
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}
	gOwner, err := strconv.Atoi(os.Getenv(OWNER_IN_OS))
	if err != nil {
		log.Panic(err)
	}

	gBot.Debug = true
	msg := tgbotapi.NewMessage(int64(gOwner), "Я снова на связи")
	gBot.Send(msg)
}

func main() {
	log.Printf("Authorized on account %s", gBot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT

	updates := gBot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			//msg.ReplyToMessageID = update.Message.MessageID

			//gBot.Send(msg)
		}
	}
}

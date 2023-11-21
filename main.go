package main

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChatState struct {
	ChatID     int64
	AllowState int
}

type QuestState struct {
	ChatID     int64
	Question   int
	AllowState int
}

var gBot *tgbotapi.BotAPI      //Указатель на бота
var gToken string              //API токен бота
var gOwner int64               //Владелец бота сюда будут приходить служебные сообщения и вопросы от бота
var gBotName string            //Имя бота, на которое бот будет отзываться в групповом чате
var gBotGender int             //Пол бота оказывает влияние на его представление
var gChatsStates []ChatState   //Для инициализации списка доступов для чатов. Сохраняется в файл
var gQuestsStates []QuestState //Для слежения за квестами в реальном времени

func SendToOwner(mesText string, quest int) {
	msg := tgbotapi.NewMessage(gOwner, mesText)
	switch quest {
	case ALLOW:
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Да", "yes"),
					tgbotapi.NewInlineKeyboardButtonData("Нет", "no"),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	}
	gBot.Send(msg)
}

func init() {
	var err error
	var owner int
	gToken = os.Getenv(TOKEN_NAME_IN_OS)
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	} else {
		gBot.Debug = true
	}
	if owner, err = strconv.Atoi(os.Getenv(OWNER_IN_OS)); err != nil {
		log.Panic(err)
	} else {
		gOwner = int64(owner)
	}
	gBotName = os.Getenv(BOTNAME_IN_OS)
	switch os.Getenv(BOTGENDER_IN_OS) {
	case "Male":
		gBotGender = MALE
	case "Female":
		gBotGender = FEMALE
	case "Neutral":
		gBotGender = NEUTRAL
	}
	gChatsStates = []ChatState{{ChatID: 0, AllowState: DISALLOW}, {ChatID: gOwner, AllowState: ALLOW}}
	SendToOwner("Я снова на связи", NOTHING)
}

func main() {
	log.Printf("Authorized on account %s", gBot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			AllowChat(update.Message.Chat.ID, update.Message.From.UserName, update.Message.Text)
			//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			//msg.ReplyToMessageID = update.Message.MessageID

			//gBot.Send(msg)
		}
	}
}

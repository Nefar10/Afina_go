package main

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SelectBotStyle(update tgbotapi.Update) {
	SetCurOperation("Style processing")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	SendToUser(gOwner, IM12[gLocale], GPTSTYLES, 1, chatID)
}

func SelectBotCharacter(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Character type")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "**Текущий Характер:**\n"+gCT[chatItem.CharType-1], SELECTCHARACTER, 1, chatItem.ChatID)
	}
}

func SelectBotModel(update tgbotapi.Update) {
	SetCurOperation("Gpt model select")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	SendToUser(gOwner, "Выберите модель"+chatIDstr, GPTSELECT, 2, chatID)
}

func SelectChat(update tgbotapi.Update) {
	var err error
	var msgString string
	var chatItem ChatState
	var keys []string
	SetCurOperation("processing callback WB lists")
	keys, err = gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		SendToUser(gOwner, E12[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	//keys processing
	msgString = ""
	for _, key := range keys {
		chatItem = GetChatStateDB(key)
		if chatItem.ChatID != 0 {
			if chatItem.AllowState == ALLOW && update.CallbackQuery.Data == "WHITELIST" {
				if chatItem.Type != "private" {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
				} else {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
				}
			}
			if (chatItem.AllowState == DISALLOW || chatItem.AllowState == BLACKLISTED) && update.CallbackQuery.Data == "BLACKLIST" {
				if chatItem.Type != "private" {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
				} else {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
				}
			}
			if chatItem.AllowState == IN_PROCESS && update.CallbackQuery.Data == "INPROCESS" {
				if chatItem.Type != "private" {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
				} else {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
				}
			}
		}
	}
	SendToUser(gOwner, msgString, SELECTCHAT, 1)
}

func SelectChatRights(update tgbotapi.Update) {
	SetCurOperation("Rights change")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	SendToUser(gOwner, "Изменить права доступа для чата "+chatIDstr, ACCESS, 2, chatID)
}

func SelectChatFacts(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat facts processing")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		SendToUser(gOwner, IM14[gLocale], INTFACTS, 1, chatItem.ChatID)
	}
}

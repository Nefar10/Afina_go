package main

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SelectBotStyle(update tgbotapi.Update) {
	SetCurOperation("Style processing", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}
	SendToUser(gOwner, 0, gIm[12][gLocale], MenuSetStyle, 1, false, chatID)
}

func SelectBotCharacter(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Character type", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		SetChatStateDB(chatItem)
		SendToUser(gOwner, 0, "**Текущий Характер:**\n"+gCharTypes[chatItem.CharType-1].Type+" "+
			gCharTypes[chatItem.CharType-1].Description[gLocale], MenuShowChar, 1, false, chatItem.ChatID)
	}
}

func SelectBotModel(update tgbotapi.Update) {
	SetCurOperation("Gpt model select", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}
	SendToUser(gOwner, 0, "Выберите модель"+chatIDstr, MenuSetModel, 2, false, chatID)
}

func SelectChat(update tgbotapi.Update) {
	var err error
	var msgString string
	var chatItem ChatState
	var keys []string
	SetCurOperation("processing callback WB lists", 0)
	keys, err = gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		SendToUser(gOwner, 0, gErr[12][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}
	//keys processing
	msgString = ""
	for _, key := range keys {
		chatItem = GetChatStateDB(ParseChatKeyID(key))
		if chatItem.ChatID != 0 {
			if chatItem.AllowState == ChatAllow && update.CallbackQuery.Data == "WHITELIST" {
				if chatItem.Type != "private" {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
				} else {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
				}
			}
			if (chatItem.AllowState == ChatDisallow || chatItem.AllowState == ChatBlacklist) && update.CallbackQuery.Data == "BLACKLIST" {
				if chatItem.Type != "private" {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
				} else {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
				}
			}
			if chatItem.AllowState == ChatInProcess && update.CallbackQuery.Data == "INPROCESS" {
				if chatItem.Type != "private" {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
				} else {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
				}
			}
		}
	}
	SendToUser(gOwner, 0, msgString, MenuSelChat, 1, false)
}

func SelectChatRights(update tgbotapi.Update) {
	SetCurOperation("Rights change", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}
	SendToUser(gOwner, 0, "разрешить общение в этом чате?"+chatIDstr, MenuGetAccess, 2, false, chatID)
}

func SelectChatFacts(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat facts processing", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		SendToUser(gOwner, 0, gIm[14][gLocale], MenuSetIf, 1, false, chatItem.ChatID)
	}
}

func SelectTimeZone(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat time zone processing", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		SendToUser(gOwner, 0, gIm[14][gLocale], MenuSetTimezone, 1, false, chatItem.ChatID)
	}
}

func DoWithChat(update tgbotapi.Update) {
	SetCurOperation("Select tuning action", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}
	SendToUser(update.CallbackQuery.From.ID, 0, "Выберите действие c чатом "+chatIDstr, MenuTuneChat, 1, false, chatID)
}

func Menu() {
	SetCurOperation("Menu show", 1)
	SendToUser(gOwner, 0, gIm[12][gLocale], MenuShowMenu, 1, false)
}

func UserMenu(update tgbotapi.Update) {
	var chatItem ChatState
	var chatID int64
	var err error
	var chatIDstr string
	SetCurOperation("User menu show", 0)
	if update.CallbackQuery != nil {
		chatIDstr = strings.Split(update.CallbackQuery.Data, " ")[1]
		chatID, err = strconv.ParseInt(chatIDstr, 10, 64)
	} else {
		chatIDstr = strconv.FormatInt(update.Message.Chat.ID, 10)
		chatID = update.Message.Chat.ID
	}
	if err != nil {
		SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		SendToUser(chatID, 0, gIm[12][gLocale], MenuShowUserMenu, 1, false)
	}
}

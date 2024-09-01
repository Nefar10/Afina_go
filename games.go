package main

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
)

func GameAlias(update tgbotapi.Update) {
	var chatItem ChatState
	var ChatMessages []openai.ChatCompletionMessage
	gCurProcName = "Game starting"
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		ChatMessages = append(ChatMessages, gHsGame[0].Prompt[gLocale]...)
		RenewDialog(chatIDstr, ChatMessages)
		SendToUser(chatID, IM16[gLocale], NOTHING, 0)
	}
}

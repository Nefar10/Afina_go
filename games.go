package main

import (
	"github.com/sashabaranov/go-openai"
)

func GameAlias(chatID int64) {
	var ChatMessages []openai.ChatCompletionMessage
	gCurProcName = "Game starting"
	ChatMessages = append(ChatMessages, gHsGame[0].Prompt[gLocale]...)
	UpdateDialog(chatID, ChatMessages)
	SendToUser(chatID, gIm[16][gLocale], NOTHING, 0)
}

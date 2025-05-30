package main

import (
	"github.com/sashabaranov/go-openai"
)

func GameAlias(chatID int64) {
	var ChatMessages []openai.ChatCompletionMessage
	SetCurOperation("Game started", 1)
	ChatMessages = append(ChatMessages, gHsGame[0].Prompt[gLocale]...)
	UpdateDialog(chatID, ChatMessages)
	SendToUser(chatID, 0, gIm[16][gLocale], MsgNothing, 0, false)
}

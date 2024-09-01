package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
)

func ProcessInitiative() {
	//Temporary variables
	var err error //Some errors
	//var jsonData []byte                             //Current json bytecode
	var chatItem ChatState                          //Current ChatState item
	var keys []string                               //Curent keys array
	var ChatMessages []openai.ChatCompletionMessage //Current prompt
	var FullPromt []openai.ChatCompletionMessage
	rd := gRand.Intn(1000) + 1
	keys, err = gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		SendToUser(gOwner, E12[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	gCurProcName = "Initiative processing"
	//keys processing
	for _, key := range keys {
		chatItem = GetChatStateDB(key)
		if chatItem.ChatID != 0 {
			if rd <= chatItem.Inity && chatItem.AllowState == ALLOW {
				act := tgbotapi.NewChatAction(chatItem.ChatID, tgbotapi.ChatTyping)
				gBot.Send(act)
				for {
					currentTime := time.Now()
					elapsedTime := currentTime.Sub(gLastRequest)
					time.Sleep(time.Second)
					if elapsedTime >= 20*time.Second && !gclient_is_busy {
						break
					}
				}
				gLastRequest = time.Now() //Прежде чем формировать запрос, запомним текущее время
				gclient_is_busy = true
				//ChatMessages = gIntFacts[chatItem.InterFacts].Prompt[gLocale]
				//currentTime := time.Now()
				//ChatMessages[len(ChatMessages)-1].Content = currentTime.Format("2006-01-02 15:04:05") + " " + ChatMessages[len(ChatMessages)-1].Content
				FullPromt = nil
				FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
				FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
				FullPromt = append(FullPromt, gIntFacts[chatItem.InterFacts].Prompt[gLocale]...)
				//log.Println(FullPromt)
				resp, err := gclient.CreateChatCompletion( //Формируем запрос к мозгам
					context.Background(),
					openai.ChatCompletionRequest{
						Model:       chatItem.Model,
						Temperature: chatItem.Temperature,
						Messages:    FullPromt,
					},
				)
				gclient_is_busy = false
				if err != nil {
					SendToUser(gOwner, E17[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, INFO, 0)
				} else {
					//log.Printf("Чат ID: %d Токенов использовано: %d", chatItem.ChatID, resp.Usage.TotalTokens)
					SendToUser(chatItem.ChatID, resp.Choices[0].Message.Content, NOTHING, 0)
				}
				ChatMessages = GetChatMessages("Dialog:" + strconv.FormatInt(chatItem.ChatID, 10))
				ChatMessages = append(ChatMessages, gIntFacts[chatItem.InterFacts].Prompt[gLocale]...)
				ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content})
				RenewDialog(strconv.FormatInt(chatItem.ChatID, 10), ChatMessages)
			}
		}
	}
}

func isMyReaction(messages []openai.ChatCompletionMessage, Bstyle int, History []openai.ChatCompletionMessage) bool {
	var FullPromt []openai.ChatCompletionMessage
	FullPromt = append(FullPromt, History...)
	if len(messages) >= 3 {
		FullPromt = append(FullPromt, messages[len(messages)-3:]...)
	} else {
		FullPromt = append(FullPromt, messages...)
	}
	FullPromt = append(FullPromt, gHsReaction[0].Prompt[gLocale]...)
	resp, err := gclient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       BASEGPTMODEL,
			Temperature: 1,
			Messages:    FullPromt,
		},
	)
	if err != nil {
		SendToUser(gOwner, E17[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, INFO, 0)
		time.Sleep(20 * time.Second)
	} else {

		if strings.Contains(resp.Choices[0].Message.Content, R1[gLocale]) {
			return true
		} else {
			return false
		}
	}
	return false
}

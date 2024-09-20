package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
		SendToUser(gOwner, gErr[12][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
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
					if elapsedTime >= 20*time.Second && !gClient_is_busy {
						break
					}
				}
				gLastRequest = time.Now() //Прежде чем формировать запрос, запомним текущее время
				gClient_is_busy = true
				FullPromt = nil
				FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
				FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
				if gRand.Intn(5) == 0 {
					FullPromt = append(FullPromt, gIntFacts[0].Prompt[gLocale][gRand.Intn(len(gIntFacts[0].Prompt[gLocale]))])
				} else {
					FullPromt = append(FullPromt, gIntFacts[chatItem.InterFacts].Prompt[gLocale][gRand.Intn(len(gIntFacts[chatItem.InterFacts].Prompt[gLocale]))])
				}
				//log.Println(FullPromt)
				resp, err := gClient.CreateChatCompletion( //Формируем запрос к мозгам
					context.Background(),
					openai.ChatCompletionRequest{
						Model:       chatItem.Model,
						Temperature: chatItem.Temperature,
						Messages:    FullPromt,
					},
				)
				gClient_is_busy = false
				if err != nil {
					SendToUser(gOwner, gErr[17][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, INFO, 0)
				} else {
					//log.Printf("Чат ID: %d Токенов использовано: %d", chatItem.ChatID, resp.Usage.TotalTokens)
					SendToUser(chatItem.ChatID, resp.Choices[0].Message.Content, NOTHING, 0)
				}
				ChatMessages = GetChatMessages("Dialog:" + strconv.FormatInt(chatItem.ChatID, 10))
				ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content})
				RenewDialog(chatItem.ChatID, ChatMessages)
			}
		}
	}
}

func isMyReaction(messages []openai.ChatCompletionMessage, History []openai.ChatCompletionMessage) bool {
	var FullPromt []openai.ChatCompletionMessage
	var resp openai.ChatCompletionResponse
	var err error
	var result bool
	result = false
	FullPromt = nil
	FullPromt = append(FullPromt, gHsName[gLocale]...)
	FullPromt = append(FullPromt, History...)
	if len(messages) >= 3 {
		FullPromt = append(FullPromt, messages[len(messages)-3:]...)
	} else {
		FullPromt = append(FullPromt, messages...)
	}
	FullPromt = append(FullPromt, gHsReaction[0].Prompt[gLocale]...)
	resp, err = gClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       BASEGPTMODEL,
			Temperature: 0,
			Messages:    FullPromt,
		},
	)
	if err != nil {
		SendToUser(gOwner, gErr[17][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, INFO, 0)
		time.Sleep(20 * time.Second)
	} else {
		log.Println(resp.Choices[0].Message.Content)
		if strings.Contains(resp.Choices[0].Message.Content, gBotReaction[0][gLocale]) {
			result = true
		}
	}
	return result
}

func needFunction(messages []openai.ChatCompletionMessage) (byte, string) {
	var FullPromt []openai.ChatCompletionMessage
	var resp openai.ChatCompletionResponse
	var err error
	var result byte
	var URI string
	var respstr []string
	FullPromt = nil
	URI = ""
	//FullPromt = append(FullPromt, CharPrmt...)
	//FullPromt = append(FullPromt, History...)
	FullPromt = append(FullPromt, messages[len(messages)-1])
	FullPromt = append(FullPromt, gHsReaction[1].Prompt[gLocale]...)
	resp, err = gClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       BASEGPTMODEL,
			Temperature: 0,
			Messages:    FullPromt,
		},
	)
	if err != nil {
		SendToUser(gOwner, gErr[17][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, INFO, 0)
		time.Sleep(20 * time.Second)
	} else {
		log.Println(resp.Choices[0].Message.Content)
		switch {
		case strings.Contains(resp.Choices[0].Message.Content, "Математика"):
			result = DOCALCULATE
		case strings.Contains(resp.Choices[0].Message.Content, "Меню"):
			result = DOSHOWMENU
		case strings.Contains(resp.Choices[0].Message.Content, "История"):
			result = DOSHOWHIST
		case strings.Contains(resp.Choices[0].Message.Content, "Чистка"):
			result = DOCLEARHIST
		case strings.Contains(resp.Choices[0].Message.Content, "Игра"):
			result = DOGAME
		case strings.Contains(resp.Choices[0].Message.Content, "Сайт"):
			respstr = strings.Split(resp.Choices[0].Message.Content, "\n")
			URI = respstr[len(respstr)-1]
			result = DOREADSITE
		default:
			result = DONOTHING
		}
	}
	return result, URI
}

func DoBotFunction(BotReaction byte, chatItem ChatState, update tgbotapi.Update, ChatMessages []openai.ChatCompletionMessage, FullPromt []openai.ChatCompletionMessage, URI string) []openai.ChatCompletionChoice {
	var resp []openai.ChatCompletionChoice
	resp = nil
	switch BotReaction {
	case DOSHOWMENU:
		{
			if update.Message.Chat.ID == gOwner {
				Menu()
			} else {
				UserMenu(update)
			}
		}
	case DOSHOWHIST:
		{
			if update.Message.From.ID == gOwner {
				sendHistory(update.Message.Chat.ID, ChatMessages)
			} else {
				SendToUser(update.Message.Chat.ID, "Извините, у вас нет доступа.", INFO, 0)
			}
		}
	case DOCLEARHIST:
		{
			SendToUser(update.Message.Chat.ID, "Извините, у вас нет доступа.", INFO, 0)
		}
	case DOGAME:
		{
			GameAlias(update.Message.Chat.ID)
		}
	case DOREADSITE:
		{
			FullPromt = append(FullPromt, ProcessWebPage(URI)...)
			gClient_is_busy = true    //Флаг занятости
			gLastRequest = time.Now() //Запомним текущее время
			resp = SendRequest(FullPromt, chatItem)
		}
	}
	return resp
}

func ProcessWebPage(URI string) []openai.ChatCompletionMessage {
	var answer []openai.ChatCompletionMessage
	answer = nil
	resp, err := http.Get(URI)
	if err != nil {
		fmt.Println("Ошибка при получении страницы:", err)
		return answer
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Ошибка при загрузке документа:", err)
		return answer
	}
	html, err := doc.Html()
	if err != nil {
		fmt.Println("Ошибка при получении HTML:", err)
		return answer
	}
	log.Println(html)
	answer = []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: html},
		{Role: openai.ChatMessageRoleUser, Content: "Проанализируй содержимое представленного контента не обращая внимания на HTML разметку. Предоставь ссылки на выбранные тобой темы."}}
	return answer
}

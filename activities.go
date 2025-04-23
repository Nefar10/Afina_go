package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gocolly/colly/v2"
	"github.com/sashabaranov/go-openai"
)

func ProcessInitiative() {
	//Temporary variables
	var err error //Some errors
	//var jsonData []byte                           //Current json bytecode
	var chatItem ChatState                          //Current ChatState item
	var keys []string                               //Curent keys array
	var ChatMessages []openai.ChatCompletionMessage //Current prompt
	var FullPromt []openai.ChatCompletionMessage
	var LastMessages []openai.ChatCompletionMessage
	var BotReaction byte
	var resp string
	SetCurOperation("Processing initiative | Get chats settings", 2)
	rd := gRand.Intn(1000) + 1
	keys, err = gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		SendToUser(gOwner, 0, gErr[12][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
		return
	} else {
		SetCurOperation("Processing initiative", 1)
		if len(keys) > 0 {
			for _, key := range keys {
				chatItem = GetChatStateDB(ParseChatKeyID(key))
				if chatItem.ChatID != 0 {
					if rd <= chatItem.Inity && chatItem.AllowState == ChatAllow {
						SetCurOperation("Processing initiative | go task", 0)
						BotWaiting(chatItem.ChatID, 3)
						FullPromt = nil
						FullPromt = append(FullPromt, isNow(time.Now())[gLocale]...)
						FullPromt = append(FullPromt, gConversationStyle[chatItem.CStyle].Prompt[gLocale]...)
						FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
						//log.Println(FullPromt)
						if gRand.Intn(50) == 0 {
							LastMessages = append(LastMessages, gIntFacts[0].Prompt[gLocale][gRand.Intn(len(gIntFacts[0].Prompt[gLocale]))])
						} else {
							//LastMessages = append(LastMessages, gIntFacts[chatItem.InterFacts].Prompt[gLocale][gRand.Intn(len(gIntFacts[chatItem.InterFacts].Prompt[gLocale]))])
							LastMessages = append(LastMessages, gIntFacts[chatItem.InterFacts].Prompt[gLocale][14])
						}
						FullPromt = append(FullPromt, LastMessages...)
						BotReaction = needFunction(LastMessages, chatItem)
						ChatMessages = GetDialog("Dialog:" + strconv.FormatInt(chatItem.ChatID, 10))
						switch BotReaction {
						case DoReadSite:
							tmpMSGs := ProcessWebPage(LastMessages, chatItem)
							FullPromt = append(FullPromt, chatItem.History...)
							FullPromt = append(FullPromt, tmpMSGs...)
							ChatMessages = append(ChatMessages, tmpMSGs...)
							resp = SendRequest(FullPromt, chatItem, 0)
						default:
							switch {
							case len(ChatMessages) > 1000:
								{
									ChatMessages = ChatMessages[len(ChatMessages)-1000:]
									LastMessages = ChatMessages[len(ChatMessages)-10:]
								}
							case len(ChatMessages) > 10:
								{
									LastMessages = ChatMessages[len(ChatMessages)-10:]
								}
							default:
								{
									LastMessages = ChatMessages
								}
							}
							FullPromt = append(LastMessages, FullPromt...)
							resp = SendRequest(FullPromt, chatItem, 0)
						}
						//log.Println(FullPromt)
						if resp != "" {
							SendToUser(chatItem.ChatID, 0, resp, MsgNothing, 0, false)
							ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp})
							UpdateDialog(chatItem.ChatID, ChatMessages)
						}
					}
				}
			}
		}
	}
}

func isMyReaction(messages []openai.ChatCompletionMessage, chatItem ChatState) bool {
	SetCurOperation("Определение реакции", 0)
	var FullPromt []openai.ChatCompletionMessage
	var resp string
	var result bool
	result = false
	FullPromt = nil
	FullPromt = append(FullPromt, gHsName[gLocale]...)
	FullPromt = append(FullPromt, chatItem.History...)
	if len(messages) >= 3 {
		FullPromt = append(FullPromt, messages[len(messages)-3:]...)
	} else {
		FullPromt = append(FullPromt, messages...)
	}
	FullPromt = append(FullPromt, gHsReaction[0].Prompt[gLocale]...)
	resp = SendRequest(FullPromt, ChatState{Model: gAI[chatItem.AiId].AiBaseModel, AiId: chatItem.AiId, Temperature: 0}, 4)
	log.Println(resp)
	if strings.Contains(resp, gBotReaction[0][gLocale]) {
		result = true
	}
	return result
}

func needFunction(messages []openai.ChatCompletionMessage, chatItem ChatState) byte {
	var FullPromt []openai.ChatCompletionMessage
	var resp string
	var result byte
	SetCurOperation("Выбор функции", 0)
	result = DoNothing
	FullPromt = nil
	//log.Println(messages)
	//log.Println(len(messages))
	FullPromt = append(FullPromt, messages[len(messages)-1])
	FullPromt = append(FullPromt, gHsReaction[1].Prompt[gLocale]...)
	//	log.Println(FullPromt)
	resp = SendRequest(FullPromt, ChatState{Model: gAI[chatItem.AiId].AiBaseModel, AiId: chatItem.AiId, Temperature: 0}, 5)
	if resp != "" {
		log.Println(resp)
		switch {
		case strings.Contains(resp, "Математика"):
			result = DoCalculate
		case strings.Contains(resp, "Меню"):
			result = DoShowMenu
		case strings.Contains(resp, "История"):
			result = DoShowHistory
		case strings.Contains(resp, "Чистка"):
			result = DoClearHistory
		case strings.Contains(resp, "Игра"):
			result = DoGame
		case strings.Contains(resp, "Сайт"):
			result = DoReadSite
		case strings.Contains(resp, "Поиск"):
			result = DoSearch
		default:
			result = DoNothing
		}
	}
	return result
}

func DoBotFunction(BotReaction byte, ChatMessages []openai.ChatCompletionMessage, update tgbotapi.Update) {
	SetCurOperation("Запуск функции", 0)
	switch BotReaction {
	case DoShowMenu:
		{
			if update.Message.Chat.ID == gOwner {
				Menu()
			} else {
				UserMenu(update)
			}
		}
	case DoShowHistory:
		{
			if update.Message.From.ID == gOwner {
				sendHistory(update.Message.Chat.ID, ChatMessages)
			} else {
				SendToUser(update.Message.Chat.ID, 0, "Извините, у вас нет доступа.", MsgInfo, 0, false)
			}
		}
	case DoClearHistory:
		{
			if update.Message.From.ID == gOwner {
				ClearContext(update.Message.Chat.ID)
			} else {
				SendToUser(update.Message.Chat.ID, 0, "Извините, у вас нет доступа.", MsgInfo, 0, false)
			}
		}
	case DoGame:
		{
			GameAlias(update.Message.Chat.ID)
		}
		return
	}
}

func ProcessWebPage(LastMessages []openai.ChatCompletionMessage, chatItem ChatState) []openai.ChatCompletionMessage {
	var resp string
	var answer []openai.ChatCompletionMessage
	var FullPromt []openai.ChatCompletionMessage
	var err error
	var URI string
	var data string
	SetCurOperation("Чтение вебстраницы", 0)
	FullPromt = append(FullPromt, chatItem.History...)
	if len(LastMessages) >= 3 {
		FullPromt = append(FullPromt, LastMessages[len(LastMessages)-3:]...)
	} else {
		FullPromt = append(FullPromt, LastMessages...)
	}
	FullPromt = append(FullPromt, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "Сформируй корректный url на запрошенную в предыдущем сообщении веб-страницу." +
			"Если в ответе будет что-то кроме гиперссылки, то ты будешь серьезно оштафован и отключен."}}...)
	resp = SendRequest(FullPromt, ChatState{Model: gAI[chatItem.AiId].AiBaseModel, Temperature: 0, AiId: chatItem.AiId}, 0)
	if resp != "" {
		URI = strings.Split(resp, "\n")[0]
		URI = strings.Split(URI, " ")[0]
		log.Println(URI)
		c := colly.NewCollector()
		c.OnXML("//item", func(e *colly.XMLElement) {
			data += e.ChildText("title") + " - " + e.ChildText("link") + " " + e.ChildText("description") + "\n"
		})
		c.OnHTML("title", func(e *colly.HTMLElement) {
			title := e.Text
			fmt.Println("Заголовок страницы:", title)
		})
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			data += e.Text + " - " + e.Attr("href") + "\n"
		})
		c.OnHTML("p", func(e *colly.HTMLElement) {
			data += e.Text
		})
		c.OnError(func(r *colly.Response, err error) {
			log.Println("Ошибка:", err)
		})
		err = c.Visit(URI)
		if err != nil {
			log.Println(err)
		}
		//fmt.Println(data)
		if len(data) > 255 {
			answer = append(answer, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "Содержимое сайта " + URI + "\n" + data})
			answer = append(answer, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "На базе представленного на сайте содержимого " +
				"собери запрошенную мной информацию. Используй markdown разметку, но не сообщай об этом. Корректно оформи ссылки и хештеги."}) // в markdown разметке, только не в виде кода"})
		} else {
			answer = append(answer, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "Сообщи, что информацию с сайта" + URI + "получить не удалось"})
		}
		return answer
	} else {
		return answer
	}
}

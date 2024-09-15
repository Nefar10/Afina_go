package main

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
)

func ProcessCallbacks(update tgbotapi.Update) {
	SetCurOperation("Callback processing", 0)
	cbData := update.CallbackQuery.Data
	switch {
	case strings.Contains(cbData, "WHITELIST") || strings.Contains(cbData, "BLACKLIST") || strings.Contains(cbData, "INPROCESS"):
		SelectChat(update)
	case strings.Contains(cbData, "RESETTODEFAULTS"):
		ResetDB()
	case strings.Contains(cbData, "FLUSHCACHE"):
		FlushCache()
	case strings.Contains(cbData, "RESTART"):
		Restart()
	case strings.Contains(cbData, "ID:"):
		DoWithChat(update)
	case strings.Contains(cbData, "CLEAR_CONTEXT:"):
		ClearContext(update)
	case strings.Contains(cbData, "GAME_IT_ALIAS"):
		GameAlias(update)
	case cbData == "MENU":
		Menu()
	case strings.Contains(cbData, "USERMENU:"):
		UserMenu(update)
	case strings.Contains(cbData, "TUNE_CHAT:"):
		SetTuneChat(update)
	case strings.Contains(cbData, "STYLE:"):
		SelectBotStyle(update)
	case strings.Contains(cbData, "_ST:"):
		SetBotStyle(update)
	case strings.Contains(update.CallbackQuery.Data, "CHAT_CHARACTER:"):
		SelectBotCharacter(update)
	case strings.Contains(update.CallbackQuery.Data, "_CT:"):
		SetBotCharacter(update)
	case strings.Contains(update.CallbackQuery.Data, "CHAT_HISTORY:"):
		SetChatHistory(update)
	case strings.Contains(update.CallbackQuery.Data, "MODEL_TEMP:"):
		SetBotTemp(update)
	case strings.Contains(update.CallbackQuery.Data, "INITIATIVE:"):
		SetBotInitiative(update)
	case strings.Contains(update.CallbackQuery.Data, "CHAT_FACTS:"):
		SelectChatFacts(update)
	case strings.Contains(update.CallbackQuery.Data, "CH_TIMEZONE:"):
		SelectTimeZone(update)
	case strings.Contains(update.CallbackQuery.Data, "_TZ:"):
		SetTimeZone(update)
	case strings.Contains(update.CallbackQuery.Data, "INFO:"):
		ShowChatInfo(update)
	case strings.Contains(update.CallbackQuery.Data, "RIGHTS:"):
		SelectChatRights(update)
	case strings.Contains(update.CallbackQuery.Data, "RIGHTS:"):
	case strings.Contains(update.CallbackQuery.Data, "GPT_MODEL:"):
		SelectBotModel(update)
	case strings.Contains(update.CallbackQuery.Data, "SEL_MODEL:"):
		SetBotModel(update)
	case strings.Contains(update.CallbackQuery.Data, "_IF:"):
		SetChatFacts(update)
	default:
		CheckChatRights(update)
	}
}

func ProcessCommand(update tgbotapi.Update) {
	command := update.Message.Command()
	switch command {
	case "menu":
		if update.Message.Chat.ID == gOwner {
			SendToUser(gOwner, gIm[12][gLocale], MENU, 1)
		} else {
			SendToUser(update.Message.Chat.ID, gIm[12][gLocale], USERMENU, 1)
		}
	case "start":
		ProcessMember(update)
	}
}

func ProcessMessage(update tgbotapi.Update) {
	var chatItem ChatState                          //Current ChatState item
	var ChatMessages []openai.ChatCompletionMessage //Current prompt
	var FullPromt []openai.ChatCompletionMessage    //Messages to send
	SetCurOperation("Update message processing", 0)
	chatItem = GetChatStateDB("ChatState:" + strconv.FormatInt(update.Message.Chat.ID, 10))
	if chatItem.ChatID != 0 && chatItem.BotState == RUN {
		switch chatItem.AllowState { //Если доступ предоставлен
		case ALLOW:
			{
				//Processing settings change
				if (gChangeSettings != gOwner || chatItem.SetState != NO_ONE) && (chatItem.ChatID == gOwner) {

					chatItem = GetChatStateDB("ChatState:" + strconv.FormatInt(gChangeSettings, 10))
					SetChatSettings(chatItem, update)
				} else {
					ChatMessages = nil                                     //Формируем новый диалог
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "") //Формирум новый ответ
					if update.Message.Chat.Type != "private" {             //Если чат не приватный, то ставим отметку - на какое соощение отвечаем
						msg.ReplyToMessageID = update.Message.MessageID
					}
					ChatMessages = GetChatMessages("Dialog:" + strconv.FormatInt(update.Message.Chat.ID, 10))
					if update.Message.Chat.Type == "private" { //Если текущий чат приватный
						ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.Text})
					} else { //Если текущи чат групповой записываем первое сообщение чата дополняя его именем текущего собеседника
						ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.From.FirstName + ": " + update.Message.Text})
					}
					RenewDialog(strconv.FormatInt(update.Message.Chat.ID, 10), ChatMessages)
					for { //Здесь мы делаем паузу, позволяющую не отправлять промпты чаще чем раз в 20 секунд
						currentTime := time.Now()
						elapsedTime := currentTime.Sub(gLastRequest)
						time.Sleep(time.Second)
						if elapsedTime >= 20*time.Second && !gClient_is_busy {
							break
						}
					}
					action := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
					toBotFlag := false
					for _, name := range gBotNames { //Определим - есть ли в контексте последнего сообщения имя бота
						if strings.Contains(strings.ToUpper(update.Message.Text), strings.ToUpper(name)) && gUpdatesQty == 0 {
							toBotFlag = true
						}
					}
					if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == gBot.Self.ID && gUpdatesQty == 0 { //Если имя бота встречается
						toBotFlag = true
						//		break
					}
					if len(ChatMessages) > 20 {
						// Удаляем первые элементы, оставляя последние 10
						ChatMessages = ChatMessages[1:]
					}
					CharPrmt := [2][]openai.ChatCompletionMessage{
						{
							{Role: openai.ChatMessageRoleUser, Content: "Important! Your personality type is " + gCT[chatItem.CharType-1]},
							//	{Role: openai.ChatMessageRoleAssistant, Content: "Understood!"},
						},
						{
							{Role: openai.ChatMessageRoleUser, Content: "Важно! Твой тип характера - " + gCT[chatItem.CharType-1]},
							//	{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"},
						},
					}
					if !toBotFlag && gUpdatesQty == 0 {
						if isMyReaction(ChatMessages, CharPrmt[gLocale], chatItem.History) {
							toBotFlag = true
						}
					}

					FullPromt = nil
					FullPromt = append(FullPromt, isNow(update, chatItem.TimeZone)[gLocale]...)
					FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
					FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
					FullPromt = append(FullPromt, CharPrmt[gLocale]...)
					if chatItem.Type != "channel" {
						FullPromt = append(FullPromt, gHsBasePrompt[gLocale]...)
					}
					FullPromt = append(FullPromt, chatItem.History...)
					FullPromt = append(FullPromt, ChatMessages...)
					//log.Println(ChatMessages)
					//log.Println("")
					//log.Println(FullPromt)
					//update.Message.Chat.Type == "private" ||
					//if strings.Contains(strings.ToUpper(update.Message.Text), strings.ToUpper("сколько")) {
					//	chatItem.Temperature = 0

					/*
						else if strings.Contains(strings.ToUpper(update.Message.Text), strings.ToUpper("сколько")) {
														ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: msg.Text})
														ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "Подумай хорошо и дай точный ответ без комментариев."})
														FullPromt = nil
														FullPromt = append(FullPromt, isNow(update, chatItem.TimeZone)[gLocale]...)
														FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
														FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
														FullPromt = append(FullPromt, CharPrmt[gLocale]...)
														if chatItem.Type != "channel" {
															FullPromt = append(FullPromt, gHsBasePrompt[gLocale]...)
														}
														FullPromt = append(FullPromt, chatItem.History...)
														FullPromt = append(FullPromt, ChatMessages...)
														time.Sleep(20 * time.Second)
													}
					*/

					if toBotFlag {
						gClient_is_busy = true
						gLastRequest = time.Now() //Прежде чем формировать запрос, запомним текущее время
						for i := 0; i < 2; i++ {
							gBot.Send(action)                          //Здесь мы продолжаем делать вид, что бот отреагировал на новое сообщение
							resp, err := gClient.CreateChatCompletion( //Формируем запрос к мозгам
								context.Background(),
								openai.ChatCompletionRequest{
									Model:       chatItem.Model,
									Temperature: chatItem.Temperature,
									Messages:    FullPromt,
								},
							)
							if err != nil {
								SendToUser(gOwner, gErr[17][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, INFO, 0)
								time.Sleep(20 * time.Second)
							} else if i == 0 && strings.Contains(strings.ToUpper(update.Message.Text), strings.ToUpper("сколько")) {
								ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content})
								ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "Подумай лучше и дай точный ответ без комментариев."})
								FullPromt = nil
								FullPromt = append(FullPromt, isNow(update, chatItem.TimeZone)[gLocale]...)
								FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
								FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
								FullPromt = append(FullPromt, CharPrmt[gLocale]...)
								if chatItem.Type != "channel" {
									FullPromt = append(FullPromt, gHsBasePrompt[gLocale]...)
								}
								FullPromt = append(FullPromt, chatItem.History...)
								FullPromt = append(FullPromt, ChatMessages...)
								time.Sleep(1 * time.Second)
							} else {
								//log.Printf("Чат ID: %d Токенов использовано: %d", update.Message.Chat.ID, resp.Usage.TotalTokens)
								msg.Text = resp.Choices[0].Message.Content //Записываем ответ в сообщение
								break
							}
						}
						gClient_is_busy = false
						ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: msg.Text})
					}
					RenewDialog(strconv.FormatInt(update.Message.Chat.ID, 10), ChatMessages)
					gBot.Send(msg)
				}
			}
		case DISALLOW:
			{
				if update.Message.Chat.Type == "private" {
					SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, 0, update.Message.Chat.ID)
				} else {
					SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться в этом чате?", ACCESS, 0, update.Message.Chat.ID)

				}
			}
		case BLACKLISTED:
			if update.Message.Chat.Type == "private" {
				log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
			} else {
				log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
			}

		case IN_PROCESS:
			{
				if update.Message.Chat.Type == "private" {
					SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, 0, update.Message.Chat.ID)
					log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
				} else {
					SendToUser(gOwner, "В группововм чате "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыли диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться в этом чате?", ACCESS, 0, update.Message.Chat.ID)
					log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
				}
			}
		}
	}
}

func ProcessMember(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat member processing", 0)
	chatItem = gDefChatState
	if update.MyChatMember != nil {
		if update.MyChatMember.NewChatMember.Status == "member" || update.MyChatMember.NewChatMember.Status == "administrator" {
			SetCurOperation("Chat initialization", 0)
			chatItem.ChatID = update.MyChatMember.Chat.ID
			chatItem.UserName = update.MyChatMember.From.UserName
			chatItem.Type = update.MyChatMember.Chat.Type
			chatItem.Title = update.MyChatMember.Chat.Title
			SetChatStateDB(chatItem)
		} else if update.MyChatMember.NewChatMember.Status == "left" {
			DestroyChat(strconv.FormatInt(update.MyChatMember.Chat.ID, 10))
			SendToUser(gOwner, "Чат был закрыт, информация о нем удалена из БД", INFO, 1)
		}
	}
	if update.Message != nil {
		if update.Message.Command() == "start" {
			SetCurOperation("Chat initialization", 0)
			chatItem.ChatID = update.Message.Chat.ID
			chatItem.UserName = update.Message.From.UserName
			chatItem.Type = update.Message.Chat.Type
			chatItem.Title = update.Message.Chat.Title
			SetChatStateDB(chatItem)
		}
	}
}

func ProcessLocation(update tgbotapi.Update) {

}

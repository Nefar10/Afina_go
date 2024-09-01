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
	SetCurOperation("Callback processing")
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
			SendToUser(gOwner, IM12[gLocale], MENU, 1)
		} else {
			SendToUser(update.Message.Chat.ID, IM12[gLocale], USERMENU, 1)
		}
	}
}

func ProcessMessage(update tgbotapi.Update) {
	var chatItem ChatState                          //Current ChatState item
	var ChatMessages []openai.ChatCompletionMessage //Current prompt
	var FullPromt []openai.ChatCompletionMessage    //Messages to send
	SetCurOperation("Update message processing")
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
						if elapsedTime >= 20*time.Second && !gclient_is_busy {
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
					if !toBotFlag && gUpdatesQty == 0 {
						if isMyReaction(ChatMessages, chatItem.Bstyle, chatItem.History) {
							toBotFlag = true
						}
					}
					if len(ChatMessages) > 20 {
						// Удаляем первые элементы, оставляя последние 10
						ChatMessages = ChatMessages[1:]
					}
					CharPrmt := [2][]openai.ChatCompletionMessage{
						{
							{Role: openai.ChatMessageRoleUser, Content: "Important! Your personality type is " + gCT[chatItem.CharType-1]},
							{Role: openai.ChatMessageRoleAssistant, Content: "Understood!"},
						},
						{
							{Role: openai.ChatMessageRoleUser, Content: "Важно! Твой тип характера - " + gCT[chatItem.CharType-1]},
							{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"},
						},
					}

					FullPromt = nil
					FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
					FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
					FullPromt = append(FullPromt, CharPrmt[gLocale]...)
					FullPromt = append(FullPromt, chatItem.History...)
					FullPromt = append(FullPromt, ChatMessages...)
					//log.Println(ChatMessages)
					//log.Println("")
					//log.Println(FullPromt)
					//update.Message.Chat.Type == "private" ||
					if toBotFlag {
						gclient_is_busy = true
						gLastRequest = time.Now() //Прежде чем формировать запрос, запомним текущее время
						for i := 0; i < 2; i++ {
							gBot.Send(action)                          //Здесь мы продолжаем делать вид, что бот отреагировал на новое сообщение
							resp, err := gclient.CreateChatCompletion( //Формируем запрос к мозгам
								context.Background(),
								openai.ChatCompletionRequest{
									Model:       chatItem.Model,
									Temperature: chatItem.Temperature,
									Messages:    FullPromt,
								},
							)
							if err != nil {
								SendToUser(gOwner, E17[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, INFO, 0)
								time.Sleep(20 * time.Second)
							} else {
								//log.Printf("Чат ID: %d Токенов использовано: %d", update.Message.Chat.ID, resp.Usage.TotalTokens)
								msg.Text = resp.Choices[0].Message.Content //Записываем ответ в сообщение
								break
							}
						}
						gclient_is_busy = false
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
	SetCurOperation("Chat member processing")
	if update.MyChatMember.NewChatMember.Status == "member" || update.MyChatMember.NewChatMember.Status == "administrator" {
		SetCurOperation("Chat initialization")
		chatItem = ChatState{
			ChatID:      update.MyChatMember.Chat.ID,
			BotState:    RUN,
			AllowState:  IN_PROCESS,
			UserName:    update.MyChatMember.From.UserName,
			Type:        update.MyChatMember.Chat.Type,
			Title:       update.MyChatMember.Chat.Title,
			Model:       BASEGPTMODEL,
			Temperature: 0.5,
			Inity:       0,
			History: append(gHsNulled[gLocale],
				openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "."},
				openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"}),
			InterFacts: 0,
			Bstyle:     0,
			SetState:   NO_ONE,
			CharType:   ESTJ,
		}
		SetChatStateDB(chatItem)
	} else if update.MyChatMember.NewChatMember.Status == "left" {
		DestroyChat(strconv.FormatInt(update.MyChatMember.Chat.ID, 10))
		SendToUser(gOwner, "Чат был закрыт, информация о нем удалена из БД", INFO, 1)
	}
}
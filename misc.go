package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func SetCurOperation(msg string) {
	gCurProcName = msg
	if gVerboseLevel > 0 {
		log.Println(msg)
	}
}

func Log(msg string, lvl byte, err error) {
	switch lvl {
	case 0:
		log.Println(msg)
	case 1:
		log.Println(msg, err)
	case 2:
		log.Fatalln(msg, err)
	}
}

func GetChatStateDB(key string) ChatState {
	var err error
	var jsonStr string
	var chatItem ChatState
	SetCurOperation("Get chat state")
	jsonStr, err = gRedisClient.Get(key).Result()
	if err == redis.Nil {
		SendToUser(gOwner, E16[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, INFO, 0)
		return ChatState{}
	} else if err != nil {
		SendToUser(gOwner, E13[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		return ChatState{}
	} else {
		err = json.Unmarshal([]byte(jsonStr), &chatItem)
		if err != nil {
			SendToUser(gOwner, E14[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		}
		return chatItem
	}
}

func SetChatStateDB(item ChatState) {
	var jsonData []byte
	var err error
	SetCurOperation("Set chat state")
	if item.CharType < 1 {
		item.CharType = ESFJ
	}
	jsonData, err = json.Marshal(item)
	if err != nil {
		SendToUser(gOwner, E11[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	} else {
		DBWrite("ChatState:"+strconv.FormatInt(item.ChatID, 10), string(jsonData), 0)
	}
}

func RenewDialog(chatIDstr string, ChatMessages []openai.ChatCompletionMessage) {
	var jsonData []byte
	var err error
	SetCurOperation("Update dialog")
	jsonData, err = json.Marshal(ChatMessages)
	if err != nil {
		SendToUser(gOwner, E11[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	DBWrite("Dialog:"+chatIDstr, string(jsonData), 0)
}

func GetChatMessages(key string) []openai.ChatCompletionMessage {
	var msgString string
	var err error
	var ChatMessages []openai.ChatCompletionMessage
	SetCurOperation("Dialog reading from DB")
	msgString, err = gRedisClient.Get(key).Result() //Пытаемся прочесть из БД диалог
	if err == redis.Nil {                           //Если диалога в БД нет, формируем новый и записываем в БД
		Log("Ошибка", ERR, err)
		return []openai.ChatCompletionMessage{}
	} else if err != nil {
		SendToUser(gOwner, E13[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		return []openai.ChatCompletionMessage{}
	} else { //Если диалог уже существует
		err = json.Unmarshal([]byte(msgString), &ChatMessages)
		if err != nil {
			SendToUser(gOwner, E14[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		}
		return ChatMessages
	}
}

func DBWrite(key string, value string, expiration time.Duration) error {
	var err = gRedisClient.Set(key, value, expiration).Err()
	if err != nil {
		SendToUser(gOwner, E10[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	return err
}

func Restart() {
	SetCurOperation("Restarting")
	SendToUser(gOwner, IM5[gLocale], INFO, 1)
	os.Exit(0)
}

func DoWithChat(update tgbotapi.Update) {
	SetCurOperation("Select tuning action")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	SendToUser(update.CallbackQuery.From.ID, "Выберите действие c чатом "+chatIDstr, TUNECHAT, 1, chatID)
}

func Menu() {
	gCurProcName = "Menu show"
	SendToUser(gOwner, IM12[gLocale], MENU, 1)
}

func UserMenu(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("User menu show")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		SendToUser(chatID, IM12[gLocale], USERMENU, 1)
	}
}

func SetTuneChat(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat tuning processing")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		SendToUser(gOwner, IM12[gLocale], TUNECHAT, 1, chatID)
	}
}

func SetBotStyle(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("GPT model changing")

	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		if strings.Contains(update.CallbackQuery.Data, "GSGOOD:") {
			chatItem.Bstyle = GOOD
			chatItem.BStPrmt = gHsGood[gLocale]
			SendToUser(gOwner, IM18[gLocale], INFO, 1)
		}
		if strings.Contains(update.CallbackQuery.Data, "GSBAD:") {
			chatItem.Bstyle = BAD
			chatItem.BStPrmt = gHsBad[gLocale]
			SendToUser(gOwner, IM19[gLocale], INFO, 1)
		}
		if strings.Contains(update.CallbackQuery.Data, "GSPOP:") {
			chatItem.Bstyle = POPPINS
			chatItem.BStPrmt = gHsPoppins[gLocale]
			SendToUser(gOwner, IM20[gLocale], INFO, 1)
		}
		if strings.Contains(update.CallbackQuery.Data, "GSSA:") {
			chatItem.Bstyle = SYSADMIN
			chatItem.BStPrmt = gHsSA[gLocale]
			SendToUser(gOwner, IM21[gLocale], INFO, 1)
		}
		SetChatStateDB(chatItem)
	}
}

func ClearContext(update tgbotapi.Update) {
	var chatItem ChatState
	var ChatMessages []openai.ChatCompletionMessage
	SetCurOperation("Context cleaning")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		ChatMessages = nil
		RenewDialog(chatIDstr, ChatMessages)
		SendToUser(chatID, "Контекст очищен!", NOTHING, 1)
	}
}

func SetBotCharacter(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Select character type")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	charValue := strings.Split(update.CallbackQuery.Data, "_")[0]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		intVal, err := strconv.Atoi(charValue)
		if err != nil {
			SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			//log.Fatalln(err, E15[gLocale]+IM29[gLocale]+gCurProcName)
		}
		chatItem.CharType = byte(intVal)
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "Выбран тип характера "+gCTDescr[gLocale][chatItem.CharType-1], INFO, 1, chatItem.ChatID)
	}
}

func SetChatHistory(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Edit history")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		chatItem.SetState = HISTORY
		gChangeSettings = chatItem.ChatID
		SetChatStateDB(chatItem)

		if len(chatItem.History) != 1 { // Патч перехода версии
			SendToUser(gOwner, "**Текущая история базовая:**\n"+chatItem.History[0].Content+"\n**Дополнитиельные факты:**\n"+chatItem.History[1].Content+"\nНапишите историю:", INFO, 1, chatItem.ChatID)
		} else {
			SendToUser(gOwner, "**Текущая история базовая:**\n"+"\n**Дополнитиельные факты:**\n"+chatItem.History[0].Content+"\nНапишите историю:", INFO, 1, chatItem.ChatID)
		}
	}
}

func SetBotTemp(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Edit temperature")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		chatItem.SetState = TEMPERATURE
		gChangeSettings = chatItem.ChatID
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "Текущий уровень - "+strconv.Itoa(int(chatItem.Temperature*10))+"\nУкажите уровень экпрессии от 1 до 10", INFO, 1, chatItem.ChatID)
	}
}

func SetBotInitiative(update tgbotapi.Update) {
	var chatItem ChatState
	gCurProcName = "Edit initiative"
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		chatItem.SetState = INITIATIVE
		gChangeSettings = chatItem.ChatID
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "Текущий уровень - "+strconv.Itoa(int(chatItem.Inity))+"\nУкажите степень инициативы от 0 до 10", INFO, 1, chatItem.ChatID)
	}
}

func ShowChatInfo(update tgbotapi.Update) {
	var msgString string
	var chatItem ChatState
	SetCurOperation("Chat info view")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		msgString = "Название чата: " + chatItem.Title + "\n" +
			"Модель поведения: " + strconv.Itoa(int(chatItem.Bstyle)) + "\n" +
			"Нейронная сеть: " + chatItem.Model + "\n" +
			"Экспрессия: " + strconv.FormatFloat(float64(chatItem.Temperature*100), 'f', -1, 32) + "%\n" +
			"Инициативность: " + strconv.Itoa(chatItem.Inity*10) + "%\n" +
			"Тип характера: " + gCTDescr[gLocale][chatItem.CharType-1] + "\n" +
			"Текущая версия: " + VER
		SendToUser(chatItem.ChatID, msgString, INFO, 2)
	}
}

func SetChatFacts(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Select chat facts")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	//log.Println(chatItem.IntFacts)
	if chatItem.ChatID != 0 {
		if strings.Contains(update.CallbackQuery.Data, "IF_GENERAL:") {
			chatItem.IntFacts = gIntFactsGen[gLocale]
		}
		if strings.Contains(update.CallbackQuery.Data, "IF_SCIENSE:") {
			chatItem.IntFacts = gIntFactsSci[gLocale]
		}
		if strings.Contains(update.CallbackQuery.Data, "IF_IT:") {
			chatItem.IntFacts = gIntFactsIT[gLocale]
		}
		if strings.Contains(update.CallbackQuery.Data, "IF_AUTO:") {
			chatItem.IntFacts = gIntFactsAuto[gLocale]
		}
		SetChatStateDB(chatItem)
		//log.Println(chatItem.IntFacts)
		SendToUser(gOwner, IM15[gLocale]+" "+chatIDstr, INFO, 1)
	}
}

func SetBotModel(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Select gpt model")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		chatItem.Model = strings.Split(update.CallbackQuery.Data, ":")[1]
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "Модель изменена на "+chatItem.Model, INFO, 1)
	}
}

func CheckChatRights(update tgbotapi.Update) {
	var err error
	var jsonStr string       //Current json string
	var questItem QuestState //Current QuestState item
	var ansItem Answer       //Curent Answer intem
	var chatItem ChatState
	SetCurOperation("Chat state changing")
	err = json.Unmarshal([]byte(update.CallbackQuery.Data), &ansItem)
	if err == nil {
		jsonStr, err = gRedisClient.Get("QuestState:" + ansItem.CallbackID.String()).Result() //читаем состояние запрса из БД
		if err == redis.Nil {
			SendToUser(gOwner, E16[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, INFO, 0)
		} else if err != nil {
			SendToUser(gOwner, E13[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		} else {
			err = json.Unmarshal([]byte(jsonStr), &questItem)
			if err != nil {
				SendToUser(gOwner, E14[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			}
			if questItem.State == QUEST_IN_PROGRESS {
				chatItem = GetChatStateDB("ChatState:" + strconv.FormatInt(questItem.ChatID, 10))
				if chatItem.ChatID != 0 {
					switch ansItem.State { //Изменяем флаг доступа
					case ALLOW:
						{
							chatItem.AllowState = ALLOW
							SendToUser(gOwner, IM6[gLocale], INFO, 0)
							SendToUser(chatItem.ChatID, IM7[gLocale], INFO, 1)
						}
					case DISALLOW:
						{
							chatItem.AllowState = DISALLOW
							SendToUser(gOwner, IM8[gLocale], INFO, 0)
							SendToUser(chatItem.ChatID, IM9[gLocale], INFO, 1)
						}
					case BLACKLISTED:
						{
							chatItem.AllowState = BLACKLISTED
							SendToUser(gOwner, IM10[gLocale], INFO, 0)
							SendToUser(chatItem.ChatID, IM11[gLocale], INFO, 1)
						}
					}
					SetChatStateDB(chatItem)
				}
			}
		}
	}
}

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
	case strings.Contains(cbData, "GSGOOD:") || strings.Contains(cbData, "GSBAD:") || strings.Contains(cbData, "GSPOP:") || strings.Contains(cbData, "GSSA:"):
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
	case strings.Contains(update.CallbackQuery.Data, "IF_"):
		SetChatFacts(update)
	default:
		CheckChatRights(update)
	}
}

func ProcessMessage(update tgbotapi.Update) {
	var err error                                   //Some errors
	var chatItem ChatState                          //Current ChatState item
	var ChatMessages []openai.ChatCompletionMessage //Current prompt
	var FullPromt []openai.ChatCompletionMessage    //Messages to send
	var temp float64
	SetCurOperation("Update message processing")
	if update.Message != nil {
		if update.Message.IsCommand() { //Begin command processing
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
		if update.Message.Text != "" { //Begin message processing
			chatItem = GetChatStateDB("ChatState:" + strconv.FormatInt(update.Message.Chat.ID, 10))
			if chatItem.ChatID != 0 {
				if chatItem.BotState == RUN {
					switch chatItem.AllowState { //Если доступ предоставлен
					case ALLOW:
						{
							//Processing settings change
							if (gChangeSettings != gOwner || chatItem.SetState != NO_ONE) && (chatItem.ChatID == gOwner) {
								chatItem = GetChatStateDB("ChatState:" + strconv.FormatInt(gChangeSettings, 10))
								if chatItem.ChatID != 0 {
									switch chatItem.SetState {
									case HISTORY:
										{
											chatItem.History = gHsNulled[gLocale]
											chatItem.History = append(chatItem.History, []openai.ChatCompletionMessage{
												{Role: openai.ChatMessageRoleUser, Content: update.Message.Text},
												{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"}}...)
										}
									case TEMPERATURE:
										{
											temp, err = strconv.ParseFloat(update.Message.Text, 64)
											if err != nil {
												SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
												//log.Fatalln(err, E15[gLocale]+IM29[gLocale]+gCurProcName)
											} else {
												chatItem.Temperature = float32(temp)
											}
											if chatItem.Temperature < 0 || chatItem.Temperature > 10 {
												chatItem.Temperature = 0.7
											} else {
												chatItem.Temperature = chatItem.Temperature / 10
											}
										}
									case INITIATIVE:
										{
											chatItem.Inity, err = strconv.Atoi(update.Message.Text)
											if err != nil {
												SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
												//log.Fatalln(err, E15[gLocale]+IM29[gLocale]+gCurProcName)
											}
											if chatItem.Inity < 0 || chatItem.Inity > 1000 {
												chatItem.Inity = 0
											}
										}
									}
									chatItem.SetState = NO_ONE
									SetChatStateDB(chatItem)
									SendToUser(gOwner, "Принято!", INFO, 1)
									gChangeSettings = gOwner
								}
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
									if isMyReaction(ChatMessages, chatItem.BStPrmt, chatItem.History) {
										toBotFlag = true
									}
								}
								if len(ChatMessages) > 20 {
									// Удаляем первые элементы, оставляя последние 10
									ChatMessages = ChatMessages[1:]
								}
								CharPrmt := [2][]openai.ChatCompletionMessage{
									{
										{Role: openai.ChatMessageRoleUser, Content: ""},
										{Role: openai.ChatMessageRoleAssistant, Content: ""},
									},
									{
										{Role: openai.ChatMessageRoleUser, Content: "Важно! Твой тип характера - " + gCT[chatItem.CharType-1]},
										{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"},
									},
								}

								FullPromt = nil
								FullPromt = append(FullPromt, chatItem.BStPrmt...)
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
						{
							if update.Message.Chat.Type == "private" {
								log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
							} else {
								log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
							}
						}
					case IN_PROCESS:
						{
							if update.Message.Chat.Type == "private" {
								log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
							} else {
								log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
							}
						}
					}
				}
			} else {
				log.Println(err) //Если записи в БД нет - формирруем новую запись
				chatItem = ChatState{
					ChatID:      update.Message.Chat.ID,
					BotState:    RUN,
					AllowState:  IN_PROCESS,
					UserName:    update.Message.From.UserName,
					Type:        update.Message.Chat.Type,
					Title:       update.Message.Chat.Title,
					Model:       BASEGPTMODEL,
					Temperature: 0.5,
					Inity:       0,
					History:     gHsNulled[gLocale],
					IntFacts:    gIntFactsGen[gLocale],
					Bstyle:      GOOD,
					BStPrmt:     gHsGood[gLocale],
					SetState:    NO_ONE,
					CharType:    ESTJ}
				SetChatStateDB(chatItem)
				if update.Message.Chat.Type == "private" {
					SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, 0, update.Message.Chat.ID)
				} else {
					SendToUser(gOwner, "В группововм чате "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыли диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться в этом чате?", ACCESS, 0, update.Message.Chat.ID)
				}
			}
		}
	}
}

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
				ChatMessages = chatItem.IntFacts
				currentTime := time.Now()
				ChatMessages[len(ChatMessages)-1].Content = currentTime.Format("2006-01-02 15:04:05") + " " + ChatMessages[len(ChatMessages)-1].Content
				FullPromt = nil
				FullPromt = append(FullPromt, chatItem.BStPrmt...)
				FullPromt = append(FullPromt, chatItem.IntFacts...)
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
				if len(ChatMessages) == 0 {
					ChatMessages = chatItem.IntFacts
				} else {
					ChatMessages = append(ChatMessages, chatItem.IntFacts...)
				}
				ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content})
				RenewDialog(strconv.FormatInt(chatItem.ChatID, 10), ChatMessages)
			}
		}
	}
}

func SendToUser(toChat int64, mesText string, quest int, ttl byte, chatID ...int64) { //отправка сообщения владельцу
	var jsonData []byte                         //Для оперативного хранения
	var jsonDataAllow []byte                    //Для формирования uuid ответа ДА
	var jsonDataDeny []byte                     //Для формирования uuid ответа НЕТ
	var jsonDataBlock []byte                    //Для формирования uuid ответа Блок
	var item QuestState                         //Для хранения состояния колбэка
	var ans Answer                              //Для формирования uuid колбэка
	msg := tgbotapi.NewMessage(toChat, mesText) //инициализируем сообщение
	SetCurOperation(IM32[gLocale])

	//Message type definition
	switch quest {
	case ERROR:
		{
			msg.Text = mesText + "\n" + IM0[gLocale]
			Log(mesText, ERR, nil)
		}
	case INFO:
		{
			msg.Text = mesText
			Log(mesText, NOERR, nil)
		}
	case ACCESS: //В случае, если стоит вопрос доступа формируем меню запроса
		{
			callbackID := uuid.New()         //создаем уникальный идентификатор запроса
			item.ChatID = chatID[0]          //указываем ID чата источника
			item.Question = quest            //указывам тип запроса
			item.CallbackID = callbackID     //запоминаем уникальнй ID
			item.State = QUEST_IN_PROGRESS   //соотояние обработки, которое запишем в БД
			item.Time = time.Now()           //запомним текущее время
			jsonData, _ = json.Marshal(item) //конвертируем структуру в json
			DBWrite("QuestState:"+callbackID.String(), string(jsonData), 0)
			ans.CallbackID = item.CallbackID //Генерируем вариант ответа "разрешить" для callback
			ans.State = ALLOW
			jsonDataAllow, _ = json.Marshal(ans) //генерируем вариант ответа "запретить" для callback
			ans.State = DISALLOW
			jsonDataDeny, _ = json.Marshal(ans) //генерируем вариант ответа "заблокировать" для callback
			ans.State = BLACKLISTED
			jsonDataBlock, _ = json.Marshal(ans)
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M1[gLocale], string(jsonDataAllow)),
					tgbotapi.NewInlineKeyboardButtonData(M2[gLocale], string(jsonDataDeny)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M3[gLocale], string(jsonDataBlock)),
				))
			msg.ReplyMarkup = numericKeyboard

		}
	case MENU: //Вызвано меню администратора
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M4[gLocale], "WHITELIST"),
					tgbotapi.NewInlineKeyboardButtonData(M5[gLocale], "BLACKLIST"),
					tgbotapi.NewInlineKeyboardButtonData(M6[gLocale], "INPROCESS"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M7[gLocale], "RESETTODEFAULTS"),
					tgbotapi.NewInlineKeyboardButtonData(M8[gLocale], "FLUSHCACHE"),
					tgbotapi.NewInlineKeyboardButtonData(M9[gLocale], "RESTART"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M10[gLocale], "TUNE_CHAT: "+strconv.FormatInt(toChat, 10)),
					tgbotapi.NewInlineKeyboardButtonData(M11[gLocale], "CLEAR_CONTEXT: "+strconv.FormatInt(toChat, 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M12[gLocale], "GAME_IT_ALIAS: "+strconv.FormatInt(toChat, 10)),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case USERMENU: //Меню подписчика
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M10[gLocale], "INFO: "+strconv.FormatInt(toChat, 10)),
					tgbotapi.NewInlineKeyboardButtonData(M11[gLocale], "CLEAR_CONTEXT: "+strconv.FormatInt(toChat, 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M12[gLocale], "GAME_IT_ALIAS: "+strconv.FormatInt(toChat, 10)),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case SELECTCHAT:
		{
			msg.Text = "Выберите чат для настройки"
			chats := strings.Split(mesText, "\n")
			var buttons []tgbotapi.InlineKeyboardButton
			for _, chat := range chats {
				if chat != "" {
					buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(chat, strings.Split(chat, "~")[0]))
				}
			}
			var rows [][]tgbotapi.InlineKeyboardButton
			for _, button := range buttons {
				row := []tgbotapi.InlineKeyboardButton{button}
				rows = append(rows, row)
			}
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.ReplyMarkup = numericKeyboard
		}
	case SELECTCHARACTER:
		{
			msg.Text = mesText
			var buttons []tgbotapi.InlineKeyboardButton
			for i := 0; i <= 15; i++ {
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(gCT[i]+" "+gCTDescr[gLocale][i], strconv.Itoa(i+1)+"_CT: "+strconv.FormatInt(chatID[0], 10)))
			}
			var rows [][]tgbotapi.InlineKeyboardButton
			for _, button := range buttons {
				row := []tgbotapi.InlineKeyboardButton{button}
				rows = append(rows, row)
			}
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.ReplyMarkup = numericKeyboard
		}
	case GPTSELECT:
		{
			msg.Text = "Выберите модель"
			var buttons []tgbotapi.InlineKeyboardButton
			for _, model := range gModels {
				if model != "" {
					buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(model, "SEL_MODEL:"+model+": "+strconv.FormatInt(chatID[0], 10)))
				}
			}

			var rows [][]tgbotapi.InlineKeyboardButton
			var row []tgbotapi.InlineKeyboardButton

			for i, button := range buttons {
				row = append(row, button)
				// Если количество кнопок в строке достигло 3, добавляем строку в rows и сбрасываем row
				if (i+1)%2 == 0 {
					rows = append(rows, row)
					row = []tgbotapi.InlineKeyboardButton{} // сброс временного среза
				}
			}

			// Если остались кнопки в последней строке, добавляем их
			if len(row) > 0 {
				rows = append(rows, row)
			}

			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.ReplyMarkup = numericKeyboard
		}
	case TUNECHAT: //меню настройки чата
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M13[gLocale], "STYLE: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(M14[gLocale], "MODEL_TEMP: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData("Нейромодель", "GPT_MODEL: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Инициатива", "INITIATIVE: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(M17[gLocale], "CHAT_FACTS: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData("Тип характера", "CHAT_CHARACTER: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M16[gLocale], "CHAT_HISTORY: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M18[gLocale], "RIGHTS: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M19[gLocale], "MENU"),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case INTFACTS: //меню настройки чата
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M20[gLocale], "IF_GENERAL: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(M21[gLocale], "IF_SCIENSE: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M22[gLocale], "IF_IT: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(M23[gLocale], "IF_AUTO: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M19[gLocale], "MENU"),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case GPTSTYLES:
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M24[gLocale], "GSGOOD: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(M25[gLocale], "GSBAD: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M27[gLocale], "GSPOP: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(M28[gLocale], "GSSA: "+strconv.FormatInt(chatID[0], 10)),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	}
	sentMsg, _ := gBot.Send(msg) //отправляем сообщение
	if ttl != 0 {
		go func() {
			time.Sleep(time.Duration(ttl) * time.Minute)
			deleteMsgConfig := tgbotapi.DeleteMessageConfig{
				ChatID:    toChat,
				MessageID: sentMsg.MessageID,
			}
			gBot.Send(deleteMsgConfig)
		}()
	}
}

func isMyReaction(messages []openai.ChatCompletionMessage, Bstyle []openai.ChatCompletionMessage, History []openai.ChatCompletionMessage) bool {
	var FullPromt []openai.ChatCompletionMessage
	//FullPromt = append(FullPromt, Bstyle...)
	FullPromt = append(FullPromt, History...)
	//FullPromt = append(FullPromt, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: mesText})
	if len(messages) >= 3 {
		FullPromt = append(FullPromt, messages[len(messages)-3:]...)
	} else {
		FullPromt = append(FullPromt, messages...)
	}
	FullPromt = append(FullPromt, gHsReaction[gLocale]...)
	//log.Println(FullPromt)
	resp, err := gclient.CreateChatCompletion( //Формируем запрос к мозгам
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       BASEGPTMODEL,
			Temperature: 1,
			Messages:    FullPromt,
		},
	)
	//log.Println(resp.Choices[0].Message.Content)
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

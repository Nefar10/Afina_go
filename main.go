package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

// test
func init() {
	//Temporary variables
	var err error
	var owner int
	var db int
	var jsonData []byte
	gRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	//Read bot API key from OS env
	gCurProcName = "Environment initialization"
	switch os.Getenv(AFINA_LOCALE_IN_OS) {
	case "Ru":
		gLocale = 1
	case "En":
		gLocale = 0
	default:
		gLocale = 0
	}
	gToken = os.Getenv(TOKEN_NAME_IN_OS)
	if gToken == "" {
		log.Fatalln(E1[gLocale] + TOKEN_NAME_IN_OS + " in process " + gCurProcName)
	}
	//Read owner's chatID from OS env
	owner, err = strconv.Atoi(os.Getenv(OWNER_IN_OS))
	if err != nil {
		log.Fatalln(err, E2[gLocale]+OWNER_IN_OS+" in process "+gCurProcName)
	} else {
		gOwner = int64(owner) //Storing owner's chat ID in variable
		gChangeSettings = gOwner
	}
	//Telegram bot init
	gBot, err = tgbotapi.NewBotAPI(gToken)
	if err != nil {
		log.Fatalln(err, E6[gLocale]+" in process "+gCurProcName)
	} else {
		//gBot.Debug = true
		log.Printf("Authorized on account %s", gBot.Self.UserName)
	}
	//Current dir init
	gDir, err = os.Getwd()
	if err != nil {
		SendToUser(gOwner, E8[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
	}
	//Read redis connector options from OS env
	//Redis IP
	gRedisIP = os.Getenv(REDIS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, E3[gLocale]+REDIS_IN_OS+" in process "+gCurProcName, ERROR, 0)
	}
	//Redis password
	gRedisPass = os.Getenv(REDIS_PASS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, E4[gLocale]+REDIS_PASS_IN_OS+" in process "+gCurProcName, ERROR, 0)
	}
	//DB ID
	db, err = strconv.Atoi(os.Getenv(REDISDB_IN_OS))
	if err != nil {
		SendToUser(gOwner, E5[gLocale]+REDISDB_IN_OS+err.Error()+" in process "+gCurProcName, ERROR, 0)
	} else {
		gRedisDB = db //Storing DB ID
	}
	//Redis client init
	gRedisClient = redis.NewClient(&redis.Options{
		Addr:     gRedisIP,
		Password: gRedisPass,
		DB:       gRedisDB,
	})
	//Chek redis connection
	err = redisPing(*gRedisClient)
	if err != nil {
		SendToUser(gOwner, E9[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
	}
	//Read OpenAI API token from OS env
	gAIToken = os.Getenv(AI_IN_OS)
	if gAIToken == "" {
		SendToUser(gOwner, E7[gLocale]+AI_IN_OS+" in process "+gCurProcName, ERROR, 0)
	}
	//Read bot names from OS env
	gBotNames = strings.Split(os.Getenv(BOTNAME_IN_OS), ",")
	if gBotNames[0] == "" {
		SendToUser(gOwner, IM1[gLocale]+BOTNAME_IN_OS+" in process "+gCurProcName, INFO, 0)
		gBotNames = []string{"Athena", "Афина"}
		log.Println(IM1[gLocale] + BOTNAME_IN_OS + " in process " + gCurProcName)
	}

	gHsName = [2][]openai.ChatCompletionMessage{
		{
			{Role: openai.ChatMessageRoleUser, Content: "Your name is " + gBotNames[0] + "."},
			{Role: openai.ChatMessageRoleAssistant, Content: "Accepted! I'm " + gBotNames[0] + "."},
		},
		{
			{Role: openai.ChatMessageRoleUser, Content: "Тебя зовут " + gBotNames[0] + ""},
			{Role: openai.ChatMessageRoleAssistant, Content: "Принято! Мое имя " + gBotNames[0] + "."},
		},
	}

	gHsGood[gLocale] = append(gHsGood[gLocale], gHsName[gLocale]...)
	gHsBad[gLocale] = append(gHsBad[gLocale], gHsName[gLocale]...)

	//Read bot gender from OS env
	switch os.Getenv(BOTGENDER_IN_OS) {
	case "Male":
		{
			gBotGender = MALE
			gHsGood[gLocale] = append(gHsGood[gLocale], gHsGenderM[gLocale]...)
			gHsBad[gLocale] = append(gHsBad[gLocale], gHsGenderM[gLocale]...)
		}
	case "Female":
		{
			gBotGender = FEMALE
			gHsGood[gLocale] = append(gHsGood[gLocale], gHsGenderF[gLocale]...)
			gHsBad[gLocale] = append(gHsBad[gLocale], gHsGenderF[gLocale]...)
		}
	case "Neutral":
		{
			gBotGender = NEUTRAL
			gHsGood[gLocale] = append(gHsGood[gLocale], gHsGenderN[gLocale]...)
			gHsBad[gLocale] = append(gHsBad[gLocale], gHsGenderN[gLocale]...)
		}
	default:
		{
			SendToUser(gOwner, IM2[gLocale]+BOTGENDER_IN_OS+" in process "+gCurProcName, INFO, 0)
			gBotGender = NEUTRAL
			gHsGood[gLocale] = append(gHsGood[gLocale], gHsGenderN[gLocale]...)
			gHsBad[gLocale] = append(gHsBad[gLocale], gHsGenderN[gLocale]...)
		}
	}
	//Default chat states init
	gChatsStates = append(gChatsStates, ChatState{ChatID: 0, Model: GPT4oMini, Inity: 0, Temperature: 0.1, AllowState: DISALLOW, UserName: "All", BotState: SLEEP, Type: "private", History: gHsNulled, IntFacts: gIntFactsGen[gLocale], BStPrmt: gHsNulled, Bstyle: GOOD, SetState: 0})
	gChatsStates = append(gChatsStates, ChatState{ChatID: gOwner, Model: GPT4oMini, Inity: 0, Temperature: 0.7, AllowState: ALLOW, UserName: "Owner", BotState: RUN, Type: "private", History: gHsNulled, IntFacts: gIntFactsGen[gLocale], BStPrmt: gHsGood[gLocale], Bstyle: GOOD, SetState: 0})
	//Storing default chat states to DB
	gCurProcName = "Chats initialization"
	for _, item := range gChatsStates {
		jsonData, err = json.Marshal(item)
		if err != nil {
			SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
		} else {
			err = gRedisClient.Set("ChatState:"+strconv.FormatInt(item.ChatID, 10), string(jsonData), 0).Err()
			if err != nil {
				SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
		}
	}
	//OpenAI client init
	config := openai.DefaultConfig(gAIToken)
	config.BaseURL = "https://api.proxyapi.ru/openai/v1"
	gclient = openai.NewClientWithConfig(config)
	gclient_is_busy = false
	//Send init complete message to owner
	SendToUser(gOwner, IM3[gLocale]+"\n"+IM13[gLocale], INFO, 0)
}

func SendToUser(toChat int64, mesText string, quest int, ttl byte, chatID ...int64) { //отправка сообщения владельцу
	var jsonData []byte                         //Для оперативного хранения
	var jsonDataAllow []byte                    //Для формирования uuid ответа ДА
	var jsonDataDeny []byte                     //Для формирования uuid ответа НЕТ
	var jsonDataBlock []byte                    //Для формирования uuid ответа Блок
	var err error                               //Временное хранение ошибок
	var item QuestState                         //Для хранения состояния колбэка
	var ans Answer                              //Для формирования uuid колбэка
	msg := tgbotapi.NewMessage(toChat, mesText) //инициализируем сообщение
	gCurProcName = "Parsing message to send"
	switch quest { //разбираем, вдруг требуется отправить запрос
	case ERROR:
		{
			msg.Text = mesText + "\n" + IM0[gLocale]
			log.Fatalln(mesText)
		}
	case INFO:
		{
			msg.Text = mesText
			log.Println(mesText)
		}
	case ACCESS: //В случае, если стоит вопрос доступа формируем меню запроса
		{
			callbackID := uuid.New()                                                             //создаем уникальный идентификатор запроса
			item.ChatID = chatID[0]                                                              //указываем ID чата источника
			item.Question = quest                                                                //указывам тип запроса
			item.CallbackID = callbackID                                                         //запоминаем уникальнй ID
			item.State = QUEST_IN_PROGRESS                                                       //соотояние обработки, которое запишем в БД
			item.Time = time.Now()                                                               //запомним текущее время
			jsonData, _ = json.Marshal(item)                                                     //конвертируем структуру в json
			err = gRedisClient.Set("QuestState:"+callbackID.String(), string(jsonData), 0).Err() //Делаем запись в БД
			if err != nil {                                                                      //Тут могут быть ошибки записи в БД
				log.Fatalln(err, E10[gLocale]+" in process "+gCurProcName)
			}
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
	case TUNECHAT: //меню настройки чата
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(M13[gLocale], "GPT_MODEL: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(M14[gLocale], "MODEL_TEMP: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(M16[gLocale], "CHAT_HISTORY: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Инициатива", "INITIATIVE: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(M17[gLocale], "CHAT_FACTS: "+strconv.FormatInt(chatID[0], 10)),
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
	FullPromt = append(FullPromt, Bstyle...)
	FullPromt = append(FullPromt, History...)
	//FullPromt = append(FullPromt, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: mesText})
	if len(messages) >= 3 {
		FullPromt = append(FullPromt, messages[len(messages)-3:]...)
	} else {
		FullPromt = append(FullPromt, messages[len(messages)-1:]...)
	}
	FullPromt = append(FullPromt, gHsReaction[gLocale]...)
	//log.Println(FullPromt)
	resp, err := gclient.CreateChatCompletion( //Формируем запрос к мозгам
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       GPT4oMini,
			Temperature: 1,
			Messages:    FullPromt,
		},
	)
	//log.Println(resp.Choices[0].Message.Content)
	if err != nil {
		SendToUser(gOwner, E17[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
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

func process_message(update tgbotapi.Update) error {
	//Temporary variables
	var err error                                   //Some errors
	var jsonStr string                              //Current json string
	var jsonData []byte                             //Current json bytecode
	var chatItem ChatState                          //Current ChatState item
	var questItem QuestState                        //Current QuestState item
	var ansItem Answer                              //Curent Answer intem
	var keys []string                               //Curent keys array
	var msgString string                            //Current message string
	var ChatMessages []openai.ChatCompletionMessage //Current prompt
	var FullPromt []openai.ChatCompletionMessage    //Messages to send
	var temp float64
	//Has been recieved callback
	//log.Println(update.CallbackQuery)
	if update.CallbackQuery != nil {
		gCurProcName = "processing callback WB lists"
		if update.CallbackQuery.Data == "WHITELIST" || update.CallbackQuery.Data == "BLACKLIST" || update.CallbackQuery.Data == "INPROCESS" {
			keys, err = gRedisClient.Keys("ChatState:*").Result()
			if err != nil {
				SendToUser(gOwner, E12[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			//keys processing
			msgString = ""
			for _, key := range keys {
				jsonStr, err = gRedisClient.Get(key).Result()
				if err == redis.Nil {
					SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
					continue
				} else if err != nil {
					SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				} else {
					err = json.Unmarshal([]byte(jsonStr), &chatItem)
					if err != nil {
						SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
					}
					if chatItem.AllowState == ALLOW && update.CallbackQuery.Data == "WHITELIST" {
						if chatItem.Type != "private" {
							msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
						} else {
							msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
						}
					}
					if chatItem.AllowState == DISALLOW && update.CallbackQuery.Data == "BLACKLIST" {
						if chatItem.Type != "private" {
							msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
						} else {
							msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
						}
					}
					if chatItem.AllowState == IN_PROCESS && update.CallbackQuery.Data == "INPROCESS" {
						if chatItem.Type != "private" {
							msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
						} else {
							msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
						}
					}
				}
			}
			SendToUser(gOwner, msgString, SELECTCHAT, 1)
		}
		gCurProcName = "Resetting"
		if update.CallbackQuery.Data == "RESETTODEFAULTS" {
			err := gRedisClient.FlushDB().Err()
			if err != nil {
				SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				SendToUser(gOwner, IM4[gLocale], INFO, 1)
				os.Exit(0)
			}
		}
		gCurProcName = "Cache cleaning"
		if update.CallbackQuery.Data == "FLUSHCACHE" {
			keys, err = gRedisClient.Keys("QuestState:*").Result()
			if err != nil {
				SendToUser(gOwner, E12[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			if len(keys) > 0 {
				msgString = "Кеш очищен\n"
				for _, key := range keys {
					err = gRedisClient.Del(key).Err()
					if err != nil {
						SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
					}
				}
				SendToUser(gOwner, msgString+"Было удалено "+strconv.Itoa(len(keys))+" устаревших записей.", INFO, 0)
			} else {
				SendToUser(gOwner, "Очищать нечего.", INFO, 1)
			}
		}
		gCurProcName = "Restarting"
		if update.CallbackQuery.Data == "RESTART" {
			SendToUser(gOwner, IM5[gLocale], INFO, 1)
			os.Exit(0)
		}
		gCurProcName = "Select tuning action"
		if strings.Contains(update.CallbackQuery.Data, "ID:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			SendToUser(update.CallbackQuery.From.ID, "Выберите действие c чатом "+chatIDstr, TUNECHAT, 1, chatID)
		}
		gCurProcName = "Context cleaning"
		if strings.Contains(update.CallbackQuery.Data, "CLEAR_CONTEXT:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result()
			if err == redis.Nil {
				SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
				return err
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				//ChatMessages = chatItem.BStPrmt
				//ChatMessages = append(ChatMessages, chatItem.History...)
				jsonData, err = json.Marshal(ChatMessages)
				if err != nil {
					SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				err = gRedisClient.Set("Dialog:"+chatIDstr, string(jsonData), 0).Err()
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				SendToUser(chatID, "Контекст очищен!", NOTHING, 1)
			}
		}
		gCurProcName = "Game starting"
		if strings.Contains(update.CallbackQuery.Data, "GAME_IT_ALIAS") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result() //Читаем инфо от чате в БД
			if err == redis.Nil {
				SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
				return err
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				ChatMessages = append(ChatMessages, gITAlias[gLocale]...)
				jsonData, err = json.Marshal(ChatMessages)
				if err != nil {
					SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				err = gRedisClient.Set("Dialog:"+chatIDstr, string(jsonData), 0).Err() //Записываем диалог в БД
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				SendToUser(chatID, IM16[gLocale], NOTHING, 0)
			}
		}
		gCurProcName = "Menu processing"
		if update.CallbackQuery.Data == "MENU" {
			SendToUser(gOwner, IM12[gLocale], MENU, 1)
		}
		if strings.Contains(update.CallbackQuery.Data, "USERMENU:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result()
			if err == redis.Nil {
				SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
				return err
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				SendToUser(chatID, IM12[gLocale], USERMENU, 1)
			}
		}
		gCurProcName = "Chat tuning processing"
		if strings.Contains(update.CallbackQuery.Data, "TUNE_CHAT:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result()
			if err == redis.Nil {
				SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
				return err
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				SendToUser(gOwner, IM12[gLocale], TUNECHAT, 1, chatID)
			}
		}
		gCurProcName = "GPT model processing"
		if strings.Contains(update.CallbackQuery.Data, "GPT_MODEL:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				log.Fatalln(err, E15[gLocale]+" in process "+gCurProcName)
			}
			SendToUser(gOwner, IM12[gLocale], GPTSTYLES, 1, chatID)
		}
		gCurProcName = "GPT model changing"
		if strings.Contains(update.CallbackQuery.Data, "GSGOOD:") || strings.Contains(update.CallbackQuery.Data, "GSBAD:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				log.Fatalln(err, E15[gLocale]+" in process "+gCurProcName)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(chatID, 10)).Result()
			if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			}
			if strings.Contains(update.CallbackQuery.Data, "GSGOOD:") {
				chatItem.Bstyle = GOOD
				chatItem.BStPrmt = gHsGood[gLocale]
				SendToUser(gOwner, IM18[gLocale], INFO, 1)
			} else {
				chatItem.Bstyle = BAD
				chatItem.BStPrmt = gHsBad[gLocale]
				SendToUser(gOwner, IM19[gLocale], INFO, 1)
			}
			jsonData, err = json.Marshal(chatItem)
			if err != nil {
				SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = gRedisClient.Set("ChatState:"+strconv.FormatInt(chatID, 10), string(jsonData), 0).Err()
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			}
		}
		gCurProcName = "Edit history"
		if strings.Contains(update.CallbackQuery.Data, "CHAT_HISTORY:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				log.Fatalln(err, E15[gLocale]+" in process "+gCurProcName)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(chatID, 10)).Result()
			if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			}
			chatItem.SetState = HISTORY
			gChangeSettings = chatID
			jsonData, err = json.Marshal(chatItem)
			if err != nil {
				SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = gRedisClient.Set("ChatState:"+strconv.FormatInt(chatID, 10), string(jsonData), 0).Err()
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			}
			SendToUser(gOwner, "Текущая история:\n"+chatItem.History[0].Content+"\nНапишите историю:", INFO, 1, chatID)
		}
		gCurProcName = "Edit temperature"
		if strings.Contains(update.CallbackQuery.Data, "MODEL_TEMP:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				log.Fatalln(err, E15[gLocale]+" in process "+gCurProcName)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(chatID, 10)).Result()
			if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			}
			chatItem.SetState = TEMPERATURE
			gChangeSettings = chatID
			jsonData, err = json.Marshal(chatItem)
			if err != nil {
				SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = gRedisClient.Set("ChatState:"+strconv.FormatInt(chatID, 10), string(jsonData), 0).Err()
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			}
			SendToUser(gOwner, "Текущий уровень - "+strconv.Itoa(int(chatItem.Temperature*10))+"\nУкажите уровень экпрессии от 1 до 10", INFO, 1, chatID)
		}
		gCurProcName = "Edit initiative"
		if strings.Contains(update.CallbackQuery.Data, "INITIATIVE:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				log.Fatalln(err, E15[gLocale]+" in process "+gCurProcName)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(chatID, 10)).Result()
			if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			}
			chatItem.SetState = INITIATIVE
			gChangeSettings = chatID
			jsonData, err = json.Marshal(chatItem)
			if err != nil {
				SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = gRedisClient.Set("ChatState:"+strconv.FormatInt(chatID, 10), string(jsonData), 0).Err()
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			}
			SendToUser(gOwner, "Текущий уровень - "+strconv.Itoa(int(chatItem.Inity))+"\nУкажите степень инициативы от 0 до 10", INFO, 1, chatID)
		}
		gCurProcName = "Chat facts processing"
		if strings.Contains(update.CallbackQuery.Data, "CHAT_FACTS:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result()
			if err == redis.Nil {
				SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
				return err
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				SendToUser(gOwner, IM14[gLocale], INTFACTS, 1, chatID)
			}
		}
		gCurProcName = "Chat info view"
		if strings.Contains(update.CallbackQuery.Data, "INFO:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result()
			if err == redis.Nil {
				SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
				return err
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				msgString = "Название чата: " + chatItem.Title + "\nМодель поведения: " + strconv.Itoa(int(chatItem.Bstyle)) + "\n" +
					"Экспрессия: " + strconv.FormatFloat(float64(chatItem.Temperature*100), 'f', -1, 32) + "%\n" +
					"Инициативность: " + strconv.Itoa(chatItem.Inity*10) + "%\n" +
					"Текущая версия: " + ver
				SendToUser(chatID, msgString, INFO, 2)
			}
		}
		gCurProcName = "Rights change"
		if strings.Contains(update.CallbackQuery.Data, "RIGHTS:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			SendToUser(gOwner, "Изменить права доступа для чата "+chatIDstr, ACCESS, 2, chatID)
		}

		gCurProcName = "Select chat facts"
		if strings.Contains(update.CallbackQuery.Data, "IF_") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			}
			jsonStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result()
			if err == redis.Nil {
				SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
				return err
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				if strings.Contains(update.CallbackQuery.Data, "IF_GENERAL:") {
					chatItem.IntFacts = gIntFactsGen[gLocale]
				}
				if strings.Contains(update.CallbackQuery.Data, "IF_SCIENCE:") {
					chatItem.IntFacts = gIntFactsSci[gLocale]
				}
				if strings.Contains(update.CallbackQuery.Data, "IF_IT:") {
					chatItem.IntFacts = gIntFactsIT[gLocale]
				}
				if strings.Contains(update.CallbackQuery.Data, "IF_AUTO:") {
					chatItem.IntFacts = gIntFactsAuto[gLocale]
				}
				jsonData, err = json.Marshal(chatItem)
				if err != nil {
					SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				err = gRedisClient.Set("ChatState:"+strconv.FormatInt(chatID, 10), string(jsonData), 0).Err()
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				SendToUser(gOwner, IM15[gLocale], INFO, 1)
			}
		}
		gCurProcName = "Chat state changing"
		err = json.Unmarshal([]byte(update.CallbackQuery.Data), &ansItem)
		if err == nil {
			jsonStr, err = gRedisClient.Get("QuestState:" + ansItem.CallbackID.String()).Result() //читаем состояние запрса из БД
			if err == redis.Nil {
				SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &questItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				if questItem.State == QUEST_IN_PROGRESS {
					jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(questItem.ChatID, 10)).Result()
					if err == redis.Nil {
						SendToUser(gOwner, E16[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
					} else if err != nil {
						SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
					} else {
						err = json.Unmarshal([]byte(jsonStr), &chatItem)
						if err != nil {
							SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
						}
					}
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
					jsonData, err = json.Marshal(chatItem)
					if err != nil {
						SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
					}
					err = gRedisClient.Set("ChatState:"+strconv.FormatInt(questItem.ChatID, 10), string(jsonData), 0).Err()
					if err != nil {
						SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
					}
				}
			}
		}
		return nil
	}
	gCurProcName = "Update message processing"
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
			return nil
		}
		if update.Message.Text != "" { //Begin message processing
			jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(update.Message.Chat.ID, 10)).Result()
			if err == redis.Nil {
				log.Println(err) //Если записи в БД нет - формирруем новую запись
				chatItem = ChatState{ChatID: update.Message.Chat.ID, BotState: RUN, AllowState: IN_PROCESS, UserName: update.Message.From.UserName,
					Type: update.Message.Chat.Type, Title: update.Message.Chat.Title, Model: GPT4oMini, Temperature: 0.7,
					Inity: 2, History: gHsNulled, IntFacts: gIntFactsGen[gLocale], Bstyle: GOOD, BStPrmt: gHsGood[gLocale], SetState: 0}
				jsonData, err = json.Marshal(chatItem)
				if err != nil {
					SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				err = gRedisClient.Set("ChatState:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err()
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				if update.Message.Chat.Type == "private" {
					SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, 0, update.Message.Chat.ID)
				} else {
					SendToUser(gOwner, "В группововм чате "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыли диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться в этом чате?", ACCESS, 0, update.Message.Chat.ID)
				}
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(jsonStr), &chatItem)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				if chatItem.BotState == RUN {
					switch chatItem.AllowState { //Если доступ предоставлен
					case ALLOW:
						{
							//Processing settings change
							if gChangeSettings != gOwner || chatItem.SetState != NO_ONE {
								jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(gChangeSettings, 10)).Result()
								if err != nil {
									SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
								} else {
									err = json.Unmarshal([]byte(jsonStr), &chatItem)
									if err != nil {
										SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
									}
								}
								switch chatItem.SetState {
								case HISTORY:
									{
										chatItem.History = []openai.ChatCompletionMessage{
											{Role: openai.ChatMessageRoleUser, Content: update.Message.Text},
											{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"},
										}
									}
								case TEMPERATURE:
									{
										temp, err = strconv.ParseFloat(update.Message.Text, 64)
										if err != nil {
											SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
											log.Fatalln(err, E15[gLocale]+" in process "+gCurProcName)
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
											SendToUser(gOwner, E15[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
											log.Fatalln(err, E15[gLocale]+" in process "+gCurProcName)
										} else {

										}
										if chatItem.Inity < 0 || chatItem.Inity > 1000 {
											chatItem.Inity = 0
										} else {

										}
									}
								default:
									{

									}
								}
								chatItem.SetState = NO_ONE
								jsonData, err = json.Marshal(chatItem)
								if err != nil {
									SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
								}
								err = gRedisClient.Set("ChatState:"+strconv.FormatInt(chatItem.ChatID, 10), string(jsonData), 0).Err()
								if err != nil {
									SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
								}
								SendToUser(gOwner, "Принято!", INFO, 1)
								gChangeSettings = gOwner

							} else {
								ChatMessages = nil                                     //Формируем новый диалог
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "") //Формирум новый ответ
								if update.Message.Chat.Type != "private" {             //Если чат не приватный, то ставим отметку - на какое соощение отвечаем
									msg.ReplyToMessageID = update.Message.MessageID
								}
								msgString, err = gRedisClient.Get("Dialog:" + strconv.FormatInt(update.Message.Chat.ID, 10)).Result() //Пытаемся прочесть из БД диалог
								if err == redis.Nil {                                                                                 //Если диалога в БД нет, формируем новый и записываем в БД
									log.Println(err)
									if update.Message.Chat.Type == "private" { //Если текущий чат приватный
										ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.Text})
									} else { //Если текущи чат групповой записываем первое сообщение чата дополняя его именем текущего собеседника
										ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.From.FirstName + ": " + update.Message.Text})
									}
									jsonData, err = json.Marshal(ChatMessages)
									if err != nil {
										SendToUser(gOwner, E11[gLocale]+err.Error(), ERROR, 0)
									}
									err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err() //Записываем диалог в БД
									if err != nil {
										SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0) //Здесь могут быть всякие ошибки записи в БД
									}
								} else if err != nil {
									SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
								} else { //Если диалог уже существует
									err = json.Unmarshal([]byte(msgString), &ChatMessages)
									if err != nil {
										SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
									}

									if update.Message.Chat.Type == "private" { //Если текущий чат приватный
										ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.Text})

									} else { //Если текущи чат групповой дописываем сообщение чата дополняя его именем текущего собеседника
										ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.From.FirstName + ": " + update.Message.Text})
									}
									jsonData, err = json.Marshal(ChatMessages)
									if err != nil {
										SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
									}
									err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err() //Записываем диалог в БД
									if err != nil {
										SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0) //Здесь могут быть всякие ошибки записи в БД
									}
								}
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
									if strings.Contains(strings.ToUpper(update.Message.Text), strings.ToUpper(name)) {
										toBotFlag = true
									}
								}
								if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == gBot.Self.ID { //Если имя бота встречается
									toBotFlag = true
									//		break
								}
								if !toBotFlag {
									if isMyReaction(ChatMessages, chatItem.BStPrmt, chatItem.History) {
										toBotFlag = true
									}
								}
								if len(ChatMessages) > 20 {
									// Удаляем первые элементы, оставляя последние 10
									ChatMessages = ChatMessages[1:]
								}
								FullPromt = nil
								FullPromt = append(FullPromt, chatItem.BStPrmt...)
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
											SendToUser(gOwner, E17[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
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
								jsonData, err = json.Marshal(ChatMessages)
								if err != nil {
									SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
								}
								err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err()
								if err != nil {
									SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
								}
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
			}
		}
	}
	return nil
}

func process_initiative() {
	//Temporary variables
	var err error                                   //Some errors
	var jsonStr string                              //Current json string
	var jsonData []byte                             //Current json bytecode
	var chatItem ChatState                          //Current ChatState item
	var keys []string                               //Curent keys array
	var msgString string                            //Current message string
	var ChatMessages []openai.ChatCompletionMessage //Current prompt
	var FullPromt []openai.ChatCompletionMessage
	rd := gRand.Intn(1000) + 1
	keys, err = gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		SendToUser(gOwner, E12[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
	}
	gCurProcName = "Initiative processing"
	//keys processing
	msgString = ""
	for _, key := range keys {
		jsonStr, err = gRedisClient.Get(key).Result()
		if err != nil {
			SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
		}
		err = json.Unmarshal([]byte(jsonStr), &chatItem)
		if err != nil {
			SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
		}
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
				SendToUser(gOwner, E17[gLocale]+err.Error()+" in process "+gCurProcName, INFO, 0)
			} else {
				//log.Printf("Чат ID: %d Токенов использовано: %d", chatItem.ChatID, resp.Usage.TotalTokens)
				SendToUser(chatItem.ChatID, resp.Choices[0].Message.Content, NOTHING, 0)

			}
			msgString, err = gRedisClient.Get("Dialog:" + strconv.FormatInt(chatItem.ChatID, 10)).Result() //Пытаемся прочесть из БД диалог
			if err == redis.Nil {                                                                          //Если диалога в БД нет, формируем новый и записываем в БД
				ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content})
				jsonData, err = json.Marshal(ChatMessages)
				if err != nil {
					SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				err = gRedisClient.Set("Dialog:"+strconv.FormatInt(chatItem.ChatID, 10), string(jsonData), 0).Err() //Записываем диалог в БД
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			} else if err != nil {
				SendToUser(gOwner, E13[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
			} else {
				err = json.Unmarshal([]byte(msgString), &ChatMessages)
				if err != nil {
					SendToUser(gOwner, E14[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				ChatMessages = append(ChatMessages, chatItem.IntFacts...)
				ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content})
				jsonData, err = json.Marshal(ChatMessages)
				if err != nil {
					SendToUser(gOwner, E11[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
				err = gRedisClient.Set("Dialog:"+strconv.FormatInt(chatItem.ChatID, 10), string(jsonData), 0).Err()
				if err != nil {
					SendToUser(gOwner, E10[gLocale]+err.Error()+" in process "+gCurProcName, ERROR, 0)
				}
			}
		}
	}
}

func main() {
	go func() {
		//Telegram update channel init
		updateConfig := tgbotapi.NewUpdate(0)
		updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
		updates := gBot.GetUpdatesChan(updateConfig)
		//Beginning of message processing
		for update := range updates {
			process_message(update)
		}
	}()
	for {
		time.Sleep(time.Minute)
		process_initiative()
	}
}

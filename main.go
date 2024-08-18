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

	//Logging level setup
	gVerboseLevel = 1

	//Randomize init
	gRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	//Read localization setting from OS env
	SetCurOperation("Environment initialization")
	switch os.Getenv(AFINA_LOCALE_IN_OS) {
	case "Ru":
		gLocale = 1
	case "En":
		gLocale = 0
	default:
		gLocale = 0
	}

	//Read bot API key from OS env
	gToken = os.Getenv(TOKEN_NAME_IN_OS)
	if gToken == "" {
		Log(E1[gLocale]+TOKEN_NAME_IN_OS+IM29[gLocale]+gCurProcName, CRIT, nil)
	}

	//Read owner's chatID from OS env
	owner, err = strconv.Atoi(os.Getenv(OWNER_IN_OS))
	if err != nil {
		Log(E2[gLocale]+OWNER_IN_OS+IM29[gLocale]+gCurProcName, CRIT, err)
	} else {
		gOwner = int64(owner) //Storing owner's chat ID in variable
		gChangeSettings = gOwner
	}

	//Telegram bot init
	gBot, err = tgbotapi.NewBotAPI(gToken)
	if err != nil {
		Log(E6[gLocale]+IM29[gLocale]+gCurProcName, CRIT, err)
	} else {
		if gVerboseLevel > 1 {
			gBot.Debug = true
		}
		Log(IM30[gLocale]+gBot.Self.UserName, NOERR, nil)
	}

	//Current dir init
	gDir, err = os.Getwd()
	if err != nil {
		SendToUser(gOwner, E8[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//Read redis connector options from OS env
	//Redis IP
	gRedisIP = os.Getenv(REDIS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, E3[gLocale]+REDIS_IN_OS+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//Redis password
	gRedisPass = os.Getenv(REDIS_PASS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, E4[gLocale]+REDIS_PASS_IN_OS+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//DB ID
	db, err = strconv.Atoi(os.Getenv(REDISDB_IN_OS))
	if err != nil {
		SendToUser(gOwner, E5[gLocale]+REDISDB_IN_OS+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
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
		SendToUser(gOwner, E9[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//Read OpenAI API token from OS env
	gAIToken = os.Getenv(AI_IN_OS)
	if gAIToken == "" {
		SendToUser(gOwner, E7[gLocale]+AI_IN_OS+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//Read bot names from OS env
	gBotNames = strings.Split(os.Getenv(BOTNAME_IN_OS), ",")
	if gBotNames[0] == "" {
		gBotNames = gDefBotNames
		SendToUser(gOwner, IM1[gLocale]+BOTNAME_IN_OS+IM29[gLocale]+gCurProcName, INFO, 0)
	}

	//Bot naming prompt
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

	//Characters completion with names
	gHsGood[gLocale] = append(gHsGood[gLocale], gHsName[gLocale]...)
	gHsBad[gLocale] = append(gHsBad[gLocale], gHsName[gLocale]...)
	gHsPoppins[gLocale] = append(gHsPoppins[gLocale], gHsName[gLocale]...)
	gHsSA[gLocale] = append(gHsSA[gLocale], gHsName[gLocale]...)

	//Read bot gender from OS env adn character comletion with gender information
	switch os.Getenv(BOTGENDER_IN_OS) {
	case "Male":
		{
			gBotGender = MALE
			gHsGood[gLocale] = append(gHsGood[gLocale], gHsGenderM[gLocale]...)
			gHsBad[gLocale] = append(gHsBad[gLocale], gHsGenderM[gLocale]...)
			gHsPoppins[gLocale] = append(gHsPoppins[gLocale], gHsGenderM[gLocale]...)
			gHsSA[gLocale] = append(gHsSA[gLocale], gHsGenderM[gLocale]...)
		}
	case "Female":
		{
			gBotGender = FEMALE
			gHsGood[gLocale] = append(gHsGood[gLocale], gHsGenderF[gLocale]...)
			gHsBad[gLocale] = append(gHsBad[gLocale], gHsGenderF[gLocale]...)
			gHsPoppins[gLocale] = append(gHsPoppins[gLocale], gHsGenderF[gLocale]...)
			gHsSA[gLocale] = append(gHsSA[gLocale], gHsGenderF[gLocale]...)
		}
	case "Neutral":
		{
			gBotGender = NEUTRAL
			gHsGood[gLocale] = append(gHsGood[gLocale], gHsGenderN[gLocale]...)
			gHsBad[gLocale] = append(gHsBad[gLocale], gHsGenderN[gLocale]...)
			gHsPoppins[gLocale] = append(gHsPoppins[gLocale], gHsGenderN[gLocale]...)
			gHsSA[gLocale] = append(gHsSA[gLocale], gHsGenderN[gLocale]...)
		}
	default:
		{
			SendToUser(gOwner, IM2[gLocale]+BOTGENDER_IN_OS+IM29[gLocale]+gCurProcName, INFO, 0)
			gBotGender = NEUTRAL
			gHsGood[gLocale] = append(gHsGood[gLocale], gHsGenderN[gLocale]...)
			gHsBad[gLocale] = append(gHsBad[gLocale], gHsGenderN[gLocale]...)
			gHsPoppins[gLocale] = append(gHsPoppins[gLocale], gHsGenderN[gLocale]...)
			gHsSA[gLocale] = append(gHsSA[gLocale], gHsGenderN[gLocale]...)
		}
	}

	//Default chat states init
	gChatsStates = append(gChatsStates, ChatState{
		ChatID:      1,
		Model:       BASEGPTMODEL,
		Inity:       0,
		Temperature: 0.1,
		AllowState:  DISALLOW,
		UserName:    "All",
		BotState:    SLEEP,
		Type:        "private",
		History:     gHsNulled[gLocale],
		IntFacts:    gIntFactsGen[gLocale],
		BStPrmt:     gHsNulled[gLocale],
		Bstyle:      GOOD,
		SetState:    NO_ONE,
		CharType:    ISTJ})
	gChatsStates = append(gChatsStates, ChatState{
		ChatID:      gOwner,
		Model:       BASEGPTMODEL,
		Inity:       0,
		Temperature: 0.5,
		AllowState:  ALLOW,
		UserName:    "Owner",
		BotState:    RUN,
		Type:        "private",
		History:     gHsNulled[gLocale],
		IntFacts:    gIntFactsGen[gLocale],
		BStPrmt:     gHsGood[gLocale],
		Bstyle:      GOOD,
		SetState:    NO_ONE,
		CharType:    ESFJ})

	//Storing default chat states to DB
	SetCurOperation(IM31[gLocale])
	for _, item := range gChatsStates {
		SetChatStateDB(item)
	}

	//OpenAI client init
	config := openai.DefaultConfig(gAIToken)
	config.BaseURL = "https://api.proxyapi.ru/openai/v1"
	gclient = openai.NewClientWithConfig(config)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	gModels = nil
	models, err := gclient.ListModels(ctx)
	if err != nil {
		SendToUser(gOwner, E18[gLocale], INFO, 1)
	} else {
		for _, model := range models.Models {
			if strings.Contains(strings.ToLower(model.ID), "gpt-4o") {
				gModels = append(gModels, model.ID)
			}
		}
	}
	gclient_is_busy = false

	//Send init complete message to owner
	SendToUser(gOwner, IM3[gLocale]+" "+IM13[gLocale], INFO, 0)
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

func process_message(update tgbotapi.Update) error {
	//Temporary variables
	var err error                                   //Some errors
	var jsonStr string                              //Current json string
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
				SendToUser(gOwner, E12[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			}
			//keys processing
			msgString = ""
			for _, key := range keys {
				chatItem = GetChatStateDB(key)
				if chatItem.ChatID != 0 {
					if chatItem.AllowState == ALLOW && update.CallbackQuery.Data == "WHITELIST" {
						if chatItem.Type != "private" {
							msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
						} else {
							msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
						}
					}
					if (chatItem.AllowState == DISALLOW || chatItem.AllowState == BLACKLISTED) && update.CallbackQuery.Data == "BLACKLIST" {
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
				SendToUser(gOwner, E10[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			} else {
				SendToUser(gOwner, IM4[gLocale], INFO, 1)
				os.Exit(0)
			}
		}
		gCurProcName = "Cache cleaning"
		if update.CallbackQuery.Data == "FLUSHCACHE" {
			keys, err = gRedisClient.Keys("QuestState:*").Result()
			if err != nil {
				SendToUser(gOwner, E12[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			}
			if len(keys) > 0 {
				msgString = "Кеш очищен\n"
				for _, key := range keys {
					err = gRedisClient.Del(key).Err()
					if err != nil {
						SendToUser(gOwner, E10[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
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
				SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			}
			SendToUser(update.CallbackQuery.From.ID, "Выберите действие c чатом "+chatIDstr, TUNECHAT, 1, chatID)
		}
		gCurProcName = "Context cleaning"
		if strings.Contains(update.CallbackQuery.Data, "CLEAR_CONTEXT:") {
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
		gCurProcName = "Game starting"
		if strings.Contains(update.CallbackQuery.Data, "GAME_IT_ALIAS") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			}
			chatItem = GetChatStateDB("ChatState:" + chatIDstr)
			if chatItem.ChatID != 0 {
				ChatMessages = append(ChatMessages, gITAlias[gLocale]...)
				RenewDialog(chatIDstr, ChatMessages)
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
				SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			}
			chatItem = GetChatStateDB("ChatState:" + chatIDstr)
			if chatItem.ChatID != 0 {
				SendToUser(chatID, IM12[gLocale], USERMENU, 1)
			}
		}
		gCurProcName = "Chat tuning processing"
		if strings.Contains(update.CallbackQuery.Data, "TUNE_CHAT:") {
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
		gCurProcName = "Style processing"
		if strings.Contains(update.CallbackQuery.Data, "STYLE:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
				//log.Fatalln(err, E15[gLocale]+IM29[gLocale]+gCurProcName)
			}
			SendToUser(gOwner, IM12[gLocale], GPTSTYLES, 1, chatID)
		}
		gCurProcName = "GPT model changing"
		if strings.Contains(update.CallbackQuery.Data, "GSGOOD:") || strings.Contains(update.CallbackQuery.Data, "GSBAD:") ||
			strings.Contains(update.CallbackQuery.Data, "GSPOP:") || strings.Contains(update.CallbackQuery.Data, "GSSA:") {
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
		SetCurOperation("Character type")
		if strings.Contains(update.CallbackQuery.Data, "CHAT_CHARACTER:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatItem = GetChatStateDB("ChatState:" + chatIDstr)
			if chatItem.ChatID != 0 {
				SetChatStateDB(chatItem)
				SendToUser(gOwner, "**Текущий Характер:**\n"+gCT[chatItem.CharType-1], SELECTCHARACTER, 1, chatItem.ChatID)
			}
		}
		SetCurOperation("Select character type")
		if strings.Contains(update.CallbackQuery.Data, "_CT:") {
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
		gCurProcName = "Edit history"
		if strings.Contains(update.CallbackQuery.Data, "CHAT_HISTORY:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatItem = GetChatStateDB("ChatState:" + chatIDstr)
			if chatItem.ChatID != 0 {
				chatItem.SetState = HISTORY
				gChangeSettings = chatItem.ChatID
				SetChatStateDB(chatItem)
				SendToUser(gOwner, "**Текущая история базовая:**\n"+chatItem.History[0].Content+"\n**Дополнитиельные факты:**\n"+chatItem.History[1].Content+"\nНапишите историю:", INFO, 1, chatItem.ChatID)
			}
		}
		gCurProcName = "Edit temperature"
		if strings.Contains(update.CallbackQuery.Data, "MODEL_TEMP:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatItem = GetChatStateDB("ChatState:" + chatIDstr)
			if chatItem.ChatID != 0 {
				chatItem.SetState = TEMPERATURE
				gChangeSettings = chatItem.ChatID
				SetChatStateDB(chatItem)
				SendToUser(gOwner, "Текущий уровень - "+strconv.Itoa(int(chatItem.Temperature*10))+"\nУкажите уровень экпрессии от 1 до 10", INFO, 1, chatItem.ChatID)
			}
		}
		gCurProcName = "Edit initiative"
		if strings.Contains(update.CallbackQuery.Data, "INITIATIVE:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatItem = GetChatStateDB("ChatState:" + chatIDstr)
			if chatItem.ChatID != 0 {
				chatItem.SetState = INITIATIVE
				gChangeSettings = chatItem.ChatID
				SetChatStateDB(chatItem)
				SendToUser(gOwner, "Текущий уровень - "+strconv.Itoa(int(chatItem.Inity))+"\nУкажите степень инициативы от 0 до 10", INFO, 1, chatItem.ChatID)
			}
		}
		gCurProcName = "Chat facts processing"
		if strings.Contains(update.CallbackQuery.Data, "CHAT_FACTS:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatItem = GetChatStateDB("ChatState:" + chatIDstr)
			if chatItem.ChatID != 0 {
				SendToUser(gOwner, IM14[gLocale], INTFACTS, 1, chatItem.ChatID)
			}
		}
		gCurProcName = "Chat info view"
		if strings.Contains(update.CallbackQuery.Data, "INFO:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatItem = GetChatStateDB("ChatState:" + chatIDstr)
			if chatItem.ChatID != 0 {
				msgString = "Название чата: " + chatItem.Title + "\nМодель поведения: " + strconv.Itoa(int(chatItem.Bstyle)) + "\n" +
					"Нейронная сеть: " + chatItem.Model + "\n" +
					"Экспрессия: " + strconv.FormatFloat(float64(chatItem.Temperature*100), 'f', -1, 32) + "%\n" +
					"Инициативность: " + strconv.Itoa(chatItem.Inity*10) + "%\n" +
					"Тип характера: " + gCTDescr[gLocale][chatItem.CharType-1] + "\n" +
					"Текущая версия: " + VER
				SendToUser(chatItem.ChatID, msgString, INFO, 2)
			}
		}
		gCurProcName = "Rights change"
		if strings.Contains(update.CallbackQuery.Data, "RIGHTS:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			}
			SendToUser(gOwner, "Изменить права доступа для чата "+chatIDstr, ACCESS, 2, chatID)
		}
		gCurProcName = "Gpt model select"
		if strings.Contains(update.CallbackQuery.Data, "GPT_MODEL:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			}
			SendToUser(gOwner, "Выберите модель"+chatIDstr, GPTSELECT, 2, chatID)
		}
		gCurProcName = "Select chat facts"
		if strings.Contains(update.CallbackQuery.Data, "IF_") {
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
		gCurProcName = "Select gpt model"
		if strings.Contains(update.CallbackQuery.Data, "SEL_MODEL:") {
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatItem = GetChatStateDB("ChatState:" + chatIDstr)
			if chatItem.ChatID != 0 {
				chatItem.Model = strings.Split(update.CallbackQuery.Data, ":")[1]
				SetChatStateDB(chatItem)
				SendToUser(gOwner, "Модель изменена на "+chatItem.Model, INFO, 1)
			}
		}
		gCurProcName = "Chat state changing"
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
										{Role: openai.ChatMessageRoleUser, Content: "Важно! Твой типа характера - " + gCT[chatItem.CharType-1]},
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
	return nil
}

func process_initiative() {
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
				CharPrmt := [2][]openai.ChatCompletionMessage{
					{
						{Role: openai.ChatMessageRoleUser, Content: ""},
						{Role: openai.ChatMessageRoleAssistant, Content: ""},
					},
					{
						{Role: openai.ChatMessageRoleUser, Content: "Важно! Твой типа характера - " + gCT[chatItem.CharType-1]},
						{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"},
					},
				}
				FullPromt = nil
				FullPromt = append(FullPromt, chatItem.BStPrmt...)
				FullPromt = append(FullPromt, CharPrmt[gLocale]...)
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

func main() {
	var updateQueue []tgbotapi.Update
	//Telegram update channel init
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)
	//Beginning of message processing
	go func() {
		for update := range updates {
			updateQueue = append(updateQueue, update) // Добавляем обновление в очередь
		}
	}()
	go func() {
		for {
			if len(updateQueue) > 0 {
				// Получаем первое обновление из очереди
				update := updateQueue[0]
				updateQueue = updateQueue[1:] // Удаляем его из очереди
				// Выводим количество оставшихся обновлений
				gUpdatesQty = len(updateQueue)
				//log.Println(strconv.Itoa(gUpdatesQty))
				// Обрабатываем обновление
				process_message(update)
			} else {
				// Ждем немного, чтобы избежать активного ожидания
				time.Sleep(3000 * time.Millisecond)
			}
		}
	}()
	for {
		time.Sleep(time.Minute)
		process_initiative()
	}
}

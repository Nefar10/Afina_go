package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

func SendToUser(toChat int64, mesText string, quest int, ttl byte, chatID ...int64) { //отправка сообщения владельцу
	var jsonData []byte                         //Для оперативного хранения
	var jsonDataAllow []byte                    //Для формирования uuid ответа ДА
	var jsonDataDeny []byte                     //Для формирования uuid ответа НЕТ
	var jsonDataBlock []byte                    //Для формирования uuid ответа Блок
	var item QuestState                         //Для хранения состояния колбэка
	var ans Answer                              //Для формирования uuid колбэка
	msg := tgbotapi.NewMessage(toChat, mesText) //инициализируем сообщение
	SetCurOperation(gIm[32][gLocale], 0)

	//Message type definition
	switch quest {
	case MSG_ERROR:
		{
			msg.Text = mesText + "\n" + gIm[0][gLocale]
			Log(mesText, ERR, nil)
		}
	case MSG_INFO:
		{
			msg.Text = mesText
			Log(mesText, NOERR, nil)
		}
	case MENU_GET_ACCESS: //В случае, если стоит вопрос доступа формируем меню запроса
		{
			callbackID := uuid.New()         //создаем уникальный идентификатор запроса
			item.ChatID = chatID[0]          //указываем ID чата источника
			item.Question = quest            //указывам тип запроса
			item.CallbackID = callbackID     //запоминаем уникальнй ID
			item.State = QUEST_IN_PROGRESS   //соотояние обработки, которое запишем в БД
			item.Time = time.Now()           //запомним текущее время
			jsonData, _ = json.Marshal(item) //конвертируем структуру в json
			DBWrite("QuestState:"+callbackID.String(), string(jsonData), 24*time.Hour)
			ans.CallbackID = item.CallbackID //Генерируем вариант ответа "разрешить" для callback
			ans.State = CHAT_ALLOW
			jsonDataAllow, _ = json.Marshal(ans) //генерируем вариант ответа "запретить" для callback
			ans.State = CHAT_DISALLOW
			jsonDataDeny, _ = json.Marshal(ans) //генерируем вариант ответа "заблокировать" для callback
			ans.State = CHAT_BLACKLIST
			jsonDataBlock, _ = json.Marshal(ans)
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[1][gLocale], string(jsonDataAllow)),
					tgbotapi.NewInlineKeyboardButtonData(gMenu[2][gLocale], string(jsonDataDeny)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[3][gLocale], string(jsonDataBlock)),
				))
			msg.ReplyMarkup = numericKeyboard

		}
	case MENU_SHOW_MENU: //Вызвано меню администратора
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[4][gLocale], "WHITELIST"),
					tgbotapi.NewInlineKeyboardButtonData(gMenu[5][gLocale], "BLACKLIST"),
					tgbotapi.NewInlineKeyboardButtonData(gMenu[6][gLocale], "INPROCESS"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[10][gLocale], "TUNE_CHAT: "+strconv.FormatInt(toChat, 10)),
					tgbotapi.NewInlineKeyboardButtonData("Информация", "INFO: "+strconv.FormatInt(toChat, 10)),
					tgbotapi.NewInlineKeyboardButtonData(gMenu[11][gLocale], "CLEAR_CONTEXT: "+strconv.FormatInt(toChat, 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[7][gLocale], "RESETTODEFAULTS"),
					tgbotapi.NewInlineKeyboardButtonData(gMenu[8][gLocale], "FLUSHCACHE"),
					tgbotapi.NewInlineKeyboardButtonData(gMenu[9][gLocale], "RESTART"),
				),

				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[12][gLocale], "GAME_IT_ALIAS: "+strconv.FormatInt(toChat, 10)),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case MENU_SHOW_USERMENU: //Меню подписчика
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Информация", "INFO: "+strconv.FormatInt(toChat, 10)),
					tgbotapi.NewInlineKeyboardButtonData(gMenu[11][gLocale], "CLEAR_CONTEXT: "+strconv.FormatInt(toChat, 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[12][gLocale], "GAME_IT_ALIAS: "+strconv.FormatInt(toChat, 10)),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case MENU_SEL_CHAT:
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
	case MENU_SHOW_CHAR:
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
	case MENU_SET_TIMEZONE:
		{
			msg.Text = mesText
			var buttons []tgbotapi.InlineKeyboardButton
			for i := 0; i <= 26; i++ {
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(gTimezones[i], strconv.Itoa(i)+"_TZ: "+strconv.FormatInt(chatID[0], 10)))
			}
			var rows [][]tgbotapi.InlineKeyboardButton
			for _, button := range buttons {
				row := []tgbotapi.InlineKeyboardButton{button}
				rows = append(rows, row)
			}
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.ReplyMarkup = numericKeyboard
		}

	case MENU_SET_MODEL:
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
	case MENU_TUNE_CHAT: //меню настройки чата
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[13][gLocale], "STYLE: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(gMenu[14][gLocale], "MODEL_TEMP: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData("Нейромодель", "GPT_MODEL: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Инициатива", "INITIATIVE: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData(gMenu[17][gLocale], "CHAT_FACTS: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData("Тип характера", "CHAT_CHARACTER: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[16][gLocale], "CHAT_HISTORY: "+strconv.FormatInt(chatID[0], 10)),
					tgbotapi.NewInlineKeyboardButtonData("Часовой пояс", "CH_TIMEZONE: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[18][gLocale], "RIGHTS: "+strconv.FormatInt(chatID[0], 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(gMenu[19][gLocale], "MENU"),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case MENU_SET_IF: //меню настройки чата
		{
			msg.Text = mesText
			var buttons []tgbotapi.InlineKeyboardButton
			for _, facts := range gIntFacts {
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(facts.Name, strconv.Itoa(int(facts.Id))+"_IF: "+strconv.FormatInt(chatID[0], 10)))
			}
			var rows [][]tgbotapi.InlineKeyboardButton
			for _, button := range buttons {
				row := []tgbotapi.InlineKeyboardButton{button}
				rows = append(rows, row)
			}
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.ReplyMarkup = numericKeyboard
		}
	case MENU_SET_STYLE:
		{
			msg.Text = mesText
			var buttons []tgbotapi.InlineKeyboardButton
			for _, style := range gConversationStyle {
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(style.Name, strconv.Itoa(int(style.Id))+"_ST: "+strconv.FormatInt(chatID[0], 10)))
			}
			var rows [][]tgbotapi.InlineKeyboardButton
			for _, button := range buttons {
				row := []tgbotapi.InlineKeyboardButton{button}
				rows = append(rows, row)
			}
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.ReplyMarkup = numericKeyboard
		}
	}
	//msg.Text = convTgmMarkdown(msg.Text)
	msg.ParseMode = "markdown"
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

func SelectBotStyle(update tgbotapi.Update) {
	SetCurOperation("Style processing", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, MSG_ERROR, 0)
	}
	SendToUser(gOwner, gIm[12][gLocale], MENU_SET_STYLE, 1, chatID)
}

func SelectBotCharacter(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Character type", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "**Текущий Характер:**\n"+gCT[chatItem.CharType-1], MENU_SHOW_CHAR, 1, chatItem.ChatID)
	}
}

func SelectBotModel(update tgbotapi.Update) {
	SetCurOperation("Gpt model select", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, MSG_ERROR, 0)
	}
	SendToUser(gOwner, "Выберите модель"+chatIDstr, MENU_SET_MODEL, 2, chatID)
}

func SelectChat(update tgbotapi.Update) {
	var err error
	var msgString string
	var chatItem ChatState
	var keys []string
	SetCurOperation("processing callback WB lists", 0)
	keys, err = gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		SendToUser(gOwner, gErr[12][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, MSG_ERROR, 0)
	}
	//keys processing
	msgString = ""
	for _, key := range keys {
		chatItem = GetChatStateDB(ParseChatKeyID(key))
		if chatItem.ChatID != 0 {
			if chatItem.AllowState == CHAT_ALLOW && update.CallbackQuery.Data == "WHITELIST" {
				if chatItem.Type != "private" {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
				} else {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
				}
			}
			if (chatItem.AllowState == CHAT_DISALLOW || chatItem.AllowState == CHAT_BLACKLIST) && update.CallbackQuery.Data == "BLACKLIST" {
				if chatItem.Type != "private" {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
				} else {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
				}
			}
			if chatItem.AllowState == CHAT_IN_PROCESS && update.CallbackQuery.Data == "INPROCESS" {
				if chatItem.Type != "private" {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
				} else {
					msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
				}
			}
		}
	}
	SendToUser(gOwner, msgString, MENU_SEL_CHAT, 1)
}

func SelectChatRights(update tgbotapi.Update) {
	SetCurOperation("Rights change", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, MSG_ERROR, 0)
	}
	SendToUser(gOwner, "разрешить общение в этом чате?"+chatIDstr, MENU_GET_ACCESS, 2, chatID)
}

func SelectChatFacts(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat facts processing", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		SendToUser(gOwner, gIm[14][gLocale], MENU_SET_IF, 1, chatItem.ChatID)
	}
}

func SelectTimeZone(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat time zone processing", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		SendToUser(gOwner, gIm[14][gLocale], MENU_SET_TIMEZONE, 1, chatItem.ChatID)
	}
}

func DoWithChat(update tgbotapi.Update) {
	SetCurOperation("Select tuning action", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, MSG_ERROR, 0)
	}
	SendToUser(update.CallbackQuery.From.ID, "Выберите действие c чатом "+chatIDstr, MENU_TUNE_CHAT, 1, chatID)
}

func Menu() {
	gCurProcName = "Menu show"
	SendToUser(gOwner, gIm[12][gLocale], MENU_SHOW_MENU, 1)
}

func UserMenu(update tgbotapi.Update) {
	var chatItem ChatState
	var chatID int64
	var err error
	var chatIDstr string
	SetCurOperation("User menu show", 0)
	if update.CallbackQuery != nil {
		chatIDstr = strings.Split(update.CallbackQuery.Data, " ")[1]
		chatID, err = strconv.ParseInt(chatIDstr, 10, 64)
	} else {
		chatIDstr = strconv.FormatInt(update.Message.Chat.ID, 10)
		chatID = update.Message.Chat.ID
	}
	if err != nil {
		SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, MSG_ERROR, 0)
	}
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		SendToUser(chatID, gIm[12][gLocale], MENU_SHOW_USERMENU, 1)
	}
}

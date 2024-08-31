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

func SelectBotStyle(update tgbotapi.Update) {
	SetCurOperation("Style processing")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	SendToUser(gOwner, IM12[gLocale], GPTSTYLES, 1, chatID)
}

func SelectBotCharacter(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Character type")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "**Текущий Характер:**\n"+gCT[chatItem.CharType-1], SELECTCHARACTER, 1, chatItem.ChatID)
	}
}

func SelectBotModel(update tgbotapi.Update) {
	SetCurOperation("Gpt model select")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	SendToUser(gOwner, "Выберите модель"+chatIDstr, GPTSELECT, 2, chatID)
}

func SelectChat(update tgbotapi.Update) {
	var err error
	var msgString string
	var chatItem ChatState
	var keys []string
	SetCurOperation("processing callback WB lists")
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

func SelectChatRights(update tgbotapi.Update) {
	SetCurOperation("Rights change")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
	if err != nil {
		SendToUser(gOwner, E15[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	SendToUser(gOwner, "Изменить права доступа для чата "+chatIDstr, ACCESS, 2, chatID)
}

func SelectChatFacts(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat facts processing")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		SendToUser(gOwner, IM14[gLocale], INTFACTS, 1, chatItem.ChatID)
	}
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

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	openai "github.com/sashabaranov/go-openai"
)

func SendToUser(toChat int64, replyTo int, mesText string, quest int, ttl byte, useAI bool, chatID ...int64) {
	var jsonData []byte      //Для оперативного хранения
	var jsonDataAllow []byte //Для формирования uuid ответа ДА
	var jsonDataDeny []byte  //Для формирования uuid ответа НЕТ
	var jsonDataBlock []byte //Для формирования uuid ответа Блок
	var item QuestState      //Для хранения состояния колбэка
	var ans Answer           //Для формирования uuid колбэка
	var ChatMessages []openai.ChatCompletionMessage
	var FullPromt []openai.ChatCompletionMessage
	var chatItem ChatState
	if useAI {
		ChatMessages = GetDialog("Dialog:" + strconv.FormatInt(toChat, 10))
		chatItem = GetChatStateDB(toChat)
		FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
		FullPromt = append(FullPromt, []openai.ChatCompletionMessage{{Role: "user",
			Content: "Перескажи в своем стиле, не упоминая об изменении стиля или пересказе: \n\n" + mesText}}...)
		//log.Println(FullPromt)
		mesText = SendRequest(FullPromt, chatItem)
		ChatMessages = append(ChatMessages, []openai.ChatCompletionMessage{{Role: "assistant", Content: mesText}}...)
		UpdateDialog(toChat, ChatMessages)
	}
	msg := tgbotapi.NewMessage(toChat, mesText) //инициализируем сообщение
	if replyTo != 0 {
		msg.ReplyToMessageID = replyTo
	}
	SetCurOperation(gIm[32][gLocale], 1)

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

			for i, model := range gModels {
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(model.AI_model_name+" - "+gAI[model.AI_ID].AI_Name, "SEL_MODEL:"+
					strconv.FormatInt(int64(i), 10)+":"+strconv.FormatInt(int64(model.AI_ID), 10)+":"+strconv.FormatInt(chatID[0], 10)))
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

	msg.Text = convTgmMarkdown(msg.Text)
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
		{
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
			}
			ClearContext(chatID)
		}
	case strings.Contains(cbData, "GAME_IT_ALIAS"):
		{
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
			}
			GameAlias(chatID)
		}
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
			SendToUser(gOwner, 0, gIm[12][gLocale], MENU_SHOW_MENU, 1, false)
		} else {
			SendToUser(update.Message.Chat.ID, 0, gIm[12][gLocale], MENU_SHOW_USERMENU, 1, false)
		}
	case "start":
		ProcessMember(update)
	}
}

func ProcessMessage(update tgbotapi.Update) {
	var BotReaction byte
	var toBotFlag bool
	var chatItem ChatState                          //Current ChatState item
	var ChatMessages []openai.ChatCompletionMessage //Current prompt
	var LastMessages []openai.ChatCompletionMessage //Current prompt
	var FullPromt []openai.ChatCompletionMessage    //Messages to send
	var resp string
	SetCurOperation("Update message processing", 0)
	chatItem = GetChatStateDB(update.Message.Chat.ID)
	if update.Message.Chat.ID == gOwner && gChangeSettings != gOwner && gChangeSettings != 0 {
		chatItem = GetChatStateDB(gChangeSettings)
		SetChatSettings(chatItem, update)
		return
	}
	if update.Message.Chat.ID == gOwner && gChangeSettings == gOwner {
		SetChatSettings(chatItem, update)
		return
	}
	if chatItem.ChatID != 0 && chatItem.BotState == BOT_RUN && chatItem.AllowState == CHAT_ALLOW { //Если доступ предоставлен
		ChatMessages = nil                                     //Формируем новый диалог
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "") //Формирум новый ответ
		if update.Message.Chat.Type != "private" {             //Если чат не приватный, то ставим отметку - на какое соощение отвечаем
			msg.ReplyToMessageID = update.Message.MessageID
		}
		ChatMessages = GetDialog("Dialog:" + strconv.FormatInt(update.Message.Chat.ID, 10))
		ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.From.FirstName + ": " + update.Message.Text})
		action := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		CharPrmt := [2][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Important! Your personality type is " + gCT[chatItem.CharType-1]},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Важно! Твой тип характера - " + gCT[chatItem.CharType-1]},
			},
		}
		//Готовим диалог
		switch {
		case len(ChatMessages) > 1000:
			{
				ChatMessages = ChatMessages[len(ChatMessages)-1000:]
				LastMessages = ChatMessages[len(ChatMessages)-20:]
			}
		case len(ChatMessages) > 20:
			{
				LastMessages = ChatMessages[len(ChatMessages)-20:]
			}
		default:
			{
				LastMessages = ChatMessages
			}
		}
		//Пытаемся понять - нужен ли ответ
		toBotFlag = false
		toBotFlag = (chatItem.Type == "private" && gUpdatesQty == 0) || (update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == gBot.Self.ID && gUpdatesQty == 0)
		for _, name := range gBotNames { //Определим - есть ли в контексте последнего сообщения имя бота
			if strings.Contains(strings.ToUpper(update.Message.Text), strings.ToUpper(name)) && gUpdatesQty == 0 {
				toBotFlag = true
			}
		}
		if !toBotFlag && gUpdatesQty == 0 {
			toBotFlag = isMyReaction(ChatMessages, chatItem)
		}
		//Определяем требуется ли выполнить функцию
		if toBotFlag {
			for {
				gBot.Send(action) //Симулируем набор текста
				currentTime := time.Now()
				elapsedTime := currentTime.Sub(gLastRequest)
				time.Sleep(time.Second)
				if elapsedTime >= 1*time.Second && !gClient_is_busy {
					break
				}
			}
			//Формируем промпт
			FullPromt = nil
			FullPromt = append(FullPromt, isNow(time.Unix(int64(update.Message.Date+((chatItem.TimeZone-15)*3600)), 0))[gLocale]...) //Текущее время
			FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)                                    //Модель поведения
			FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)                                                  //Пол
			FullPromt = append(FullPromt, CharPrmt[gLocale]...)                                                                      //Стиль общения
			if chatItem.Type != "channel" && chatItem.Type != "private" {
				FullPromt = append(FullPromt, gHsBasePrompt[0].Prompt[gLocale]...) //Включить базовый промпт для группы
			}
			FullPromt = append(FullPromt, chatItem.History...) //История группы
			FullPromt = append(FullPromt, LastMessages...)     //Последние сообщения
			BotReaction = needFunction(LastMessages, chatItem)
			switch BotReaction {
			case DOCALCULATE:
				{
					resp = SendRequest(FullPromt, chatItem)
					if resp != "" {
						ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp})
						ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{
							Role: openai.ChatMessageRoleUser, Content: "Найди способ решить это корректно. Дай ответ в своем стиле, не комментируя свою ошибку."})
						FullPromt = nil
						FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
						FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
						FullPromt = append(FullPromt, CharPrmt[gLocale]...)
						if len(LastMessages) >= 4 {
							FullPromt = append(FullPromt, ChatMessages[len(ChatMessages)-4:]...)
						} else {
							FullPromt = append(FullPromt, ChatMessages...)
						}

						time.Sleep(5 * time.Second)
						resp = SendRequest(FullPromt, chatItem)
					}
				}
			case DOSHOWMENU, DOSHOWHIST, DOCLEARHIST, DOGAME:
				DoBotFunction(BotReaction, ChatMessages, update)
			case DOREADSITE:
				tmpMSGs := ProcessWebPage(LastMessages, chatItem)
				FullPromt = append(FullPromt, tmpMSGs...)
				ChatMessages = append(ChatMessages, tmpMSGs...)
				resp = SendRequest(FullPromt, chatItem)
			default:
				resp = SendRequest(FullPromt, chatItem)
			}
			if (BotReaction <= DOCALCULATE || BotReaction == DOREADSITE) && resp != "" {
				msg.Text = resp
				ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: msg.Text})
				UpdateDialog(update.Message.Chat.ID, ChatMessages)
				msg.Text = convTgmMarkdown(msg.Text)
				msg.ParseMode = "markdown"
				_, err := gBot.Send(msg)
				if err != nil {
					log.Printf("Ошибка при отправке сообщения: %v", err)
					return
				}

				log.Println("Сообщение отправлено успешно")
				if chatItem.Title != update.Message.Chat.Title {
					chatItem.Title = update.Message.Chat.Title
					SetChatStateDB(chatItem)
				}
			}
		}
		return
	}
	//Обработаем иные состояния чата
	switch chatItem.AllowState {
	case CHAT_DISALLOW:
		{
			if update.Message.Chat.Type == "private" {
				SendToUser(gOwner, 0, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", MENU_GET_ACCESS, 0, false, update.Message.Chat.ID)
			} else {
				SendToUser(gOwner, 0, "Пользователь "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться в этом чате?", MENU_GET_ACCESS, 0, false, update.Message.Chat.ID)

			}
		}
	case CHAT_BLACKLIST:
		if update.Message.Chat.Type == "private" {
			log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
		} else {
			log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
		}

	case CHAT_IN_PROCESS:
		{
			if update.Message.Chat.Type == "private" {
				SendToUser(gOwner, 0, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", MENU_GET_ACCESS, 0, false, update.Message.Chat.ID)
				log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
				ProcessMember(update)
			} else {
				SendToUser(gOwner, 0, "В группововм чате "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыли диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться в этом чате?", MENU_GET_ACCESS, 0, false, update.Message.Chat.ID)
				log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
			}
		}

	}
}

func ProcessMember(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat member processing", 0)
	chatItem = gDefChatState
	chatItem.Model = gAI[chatItem.AI_ID].AI_BaseModel
	if update.MyChatMember != nil {
		if update.MyChatMember.NewChatMember.Status == "member" || update.MyChatMember.NewChatMember.Status == "administrator" {
			SetCurOperation("Chat initialization", 1)
			chatItem.ChatID = update.MyChatMember.Chat.ID
			chatItem.UserName = update.MyChatMember.From.UserName
			chatItem.Type = update.MyChatMember.Chat.Type
			chatItem.Title = update.MyChatMember.Chat.Title
			SetChatStateDB(chatItem)
		} else if update.MyChatMember.NewChatMember.Status == "left" {
			DestroyChat(strconv.FormatInt(update.MyChatMember.Chat.ID, 10))
			SendToUser(gOwner, 0, "Чат был закрыт, информация о нем удалена из БД", MSG_INFO, 1, false)
		}
	}
	if update.Message != nil {
		if update.Message.Command() == "start" {
			SetCurOperation("Chat initialization", 1)
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

func encodeImage(imagePath string) (string, error) {
	imageFile, err := os.ReadFile(imagePath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(imageFile), nil
}

func ProcessPhoto(update tgbotapi.Update) {
	var fileID string
	var fileURL string
	var caption string
	var ChatMessages []openai.ChatCompletionMessage
	//	var FullPromt []openai.ChatCompletionMessage
	var chatItem ChatState
	fileID = update.Message.Photo[len(update.Message.Photo)-1].FileID
	//fileID = update.Message.Photo[0].FileID
	caption = update.Message.Caption + "Максимально подробно опиши содержание изображения. Распознай и верни весь текст"
	fileURL = GetFileURL(fileID, update)
	ChatMessages = GetDialog("Dialog:" + strconv.FormatInt(update.Message.Chat.ID, 10))
	chatItem = GetChatStateDB(update.Message.Chat.ID)
	//log.Println(gConversationStyle[chatItem.Bstyle].Prompt[gLocale][0].Content + "\n\n" + caption)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r, err := gClient[0].CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       gAI[0].AI_BaseModel,
			Temperature: chatItem.Temperature,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: "text",
							Text: caption,
						},
						{
							Type: "image_url",
							ImageURL: &openai.ChatMessageImageURL{
								URL: fileURL,
							},
						},
					},
				},
			},
		},
	)
	log.Println(r, err)
	ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: "assistant", Content: "Получено изображение от " + update.Message.From.FirstName +
		" со следующим содержимым :" + r.Choices[0].Message.Content})
	UpdateDialog(update.Message.Chat.ID, ChatMessages)
	if (update.Message.Chat.Type == "private") && gUpdatesQty == 0 {
		if update.Message.Caption != "" {
			SendToUser(update.Message.Chat.ID, 0, r.Choices[0].Message.Content, MSG_INFO, 0, true)
		} else {
			SendToUser(update.Message.Chat.ID, 0, "Получено изображение. Что необходимо с ним сделать?", MSG_INFO, 0, true)
		}
		return
	}
	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == gBot.Self.ID && gUpdatesQty == 0 {
		if update.Message.Caption != "" {
			SendToUser(update.Message.Chat.ID, update.Message.MessageID, r.Choices[0].Message.Content, MSG_INFO, 0, true)
		} else {
			SendToUser(update.Message.Chat.ID, update.Message.MessageID, "Получено изображение. Что необходимо с ним сделать?", MSG_INFO, 0, true)
		}
		return
	}
}

func ProcessDocument(update tgbotapi.Update) {
	var fileID string
	var filePath string
	var fullFilePath string
	var fileURL string
	var realName string
	var err error
	var ChatMessages []openai.ChatCompletionMessage
	fileID = update.Message.Document.FileID
	realName = update.Message.Document.FileName
	if update.Message.Document.FileSize > 10*1024*1024 {
		if update.Message.Chat.Type == "private" {
			SendToUser(update.Message.Chat.ID, 0, "Файл "+realName+" был получен, но не будет обработан, т.к. превышает допустимый размер 10 Мб", MSG_ERROR, 1, true)
		}
		return
	}
	fileURL = GetFileURL(fileID, update)
	filePath = fmt.Sprintf("./downloads/%s", strconv.FormatInt(update.Message.Chat.ID, 10))
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		if update.Message.Chat.Type == "private" {
			SendToUser(update.Message.Chat.ID, 0, "Не удалось найти место для размещения файла "+err.Error(), MSG_ERROR, 1, true)
		}
		return
	}
	fullFilePath = fmt.Sprintf("%s/%s", filePath, fileID)
	err = downloadFile(fileURL, fullFilePath)
	if err != nil {
		if update.Message.Chat.Type == "private" {
			SendToUser(update.Message.Chat.ID, 0, "Во время загрузки файла произошла ошибка "+err.Error(), MSG_ERROR, 1, true)
		}
		return
	}
	tfile, err := os.Open(fullFilePath)
	if err != nil {
		if update.Message.Chat.Type == "private" {
			SendToUser(update.Message.Chat.ID, 0, "Файл был получен, но его не удается прочитать "+err.Error(), MSG_ERROR, 1, true)
		}
		return
	}
	defer tfile.Close()
	tfile.Seek(0, 0)
	kind, err := filetype.MatchReader(tfile)
	if err != nil {
		if update.Message.Chat.Type == "private" {
			SendToUser(update.Message.Chat.ID, 0, "Возникла ошибка при попытке определения типа файла "+err.Error(), MSG_ERROR, 1, true)
		}
		return

	}
	ChatMessages = GetDialog("Dialog:" + strconv.FormatInt(update.Message.Chat.ID, 10))
	DBWrite("File:"+fileID, fmt.Sprintf("%s/%s", filePath, realName), 0)
	switch kind.MIME.Type {
	default:
		if (update.Message.Chat.Type == "private") && !strings.HasSuffix(realName, ".txt") {
			SendToUser(update.Message.Chat.ID, 0, "Обработка формата полученного файла еще не предусмотрена.", MSG_ERROR, 1, true)
			os.Remove(filePath)
			return
		}
		content, err := os.ReadFile(fullFilePath)
		if err != nil {
			SendToUser(update.Message.Chat.ID, 0, "Не удается прочитать содержимое файла "+err.Error(), MSG_ERROR, 1, true)
			os.Remove(filePath)
			return
		}

		ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: "user", Content: "Содержимое документа " +
			update.Message.Document.FileName + "от " + update.Message.From.FirstName + "\n\n" + string(content)})
		UpdateDialog(update.Message.Chat.ID, ChatMessages)
		log.Println(ChatMessages)
		if update.Message.Chat.Type == "private" && gUpdatesQty == 0 {
			SendToUser(update.Message.Chat.ID, 0, "Что необходимо сделать с полученным документом?", MSG_INFO, 1, true)
			return
		}
		if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == gBot.Self.ID && gUpdatesQty == 0 {
			SendToUser(update.Message.Chat.ID, update.Message.MessageID, "Что необходимо сделать с полученным документом?", MSG_INFO, 0, true)
			return
		}
	}
}

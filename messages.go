package main

import (
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
		{
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
			}
			ClearContext(chatID)
		}
	case strings.Contains(cbData, "GAME_IT_ALIAS"):
		{
			chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
			chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
			if err != nil {
				SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
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
			SendToUser(gOwner, gIm[12][gLocale], MENU, 1)
		} else {
			SendToUser(update.Message.Chat.ID, gIm[12][gLocale], USERMENU, 1)
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
	var resp openai.ChatCompletionResponse
	SetCurOperation("Update message processing", 0)
	//Получим информацию о чате
	chatItem = GetChatStateDB("ChatState:" + strconv.FormatInt(update.Message.Chat.ID, 10))
	//Проверим - не требуется ли настройка
	if update.Message.Chat.ID == gOwner && gChangeSettings != gOwner && gChangeSettings != 0 {
		chatItem = GetChatStateDB("ChatState:" + strconv.FormatInt(gChangeSettings, 10))
		SetChatSettings(chatItem, update)
		return
	}
	if update.Message.Chat.ID == gOwner && gChangeSettings == gOwner {
		SetChatSettings(chatItem, update)
		return
	}
	//Если бот в работе, обработаем сообщение
	if chatItem.ChatID != 0 && chatItem.BotState == RUN && chatItem.AllowState == ALLOW { //Если доступ предоставлен
		ChatMessages = nil                                     //Формируем новый диалог
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "") //Формирум новый ответ
		if update.Message.Chat.Type != "private" {             //Если чат не приватный, то ставим отметку - на какое соощение отвечаем
			msg.ReplyToMessageID = update.Message.MessageID
		}
		ChatMessages = GetChatMessages("Dialog:" + strconv.FormatInt(update.Message.Chat.ID, 10))
		ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.From.FirstName + ": " + update.Message.Text})
		//Симулируем набор текста
		action := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		//Моделим тип личности
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
			toBotFlag = isMyReaction(ChatMessages, chatItem.History)
		}
		//Определяем требуется ли выполнить функцию
		if toBotFlag {
			for {
				gBot.Send(action) //Симулируем набор текста
				currentTime := time.Now()
				elapsedTime := currentTime.Sub(gLastRequest)
				time.Sleep(time.Second)
				if elapsedTime >= 10*time.Second && !gClient_is_busy {
					break
				}
			}
			//Формируем промпт
			FullPromt = nil
			FullPromt = append(FullPromt, isNow(update, chatItem.TimeZone)[gLocale]...)           //Текущее время
			FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...) //Модель поведения
			FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)               //Пол
			FullPromt = append(FullPromt, CharPrmt[gLocale]...)                                   //Стиль общения
			if chatItem.Type != "channel" && chatItem.Type != "private" {
				FullPromt = append(FullPromt, gHsBasePrompt[gLocale]...) //Включить базовый промпт для группы
			}
			FullPromt = append(FullPromt, chatItem.History...) //История группы
			FullPromt = append(FullPromt, LastMessages...)     //Последние сообщения
			BotReaction = needFunction(LastMessages)
			switch BotReaction {
			case DOCALCULATE:
				{
					resp = SendRequest(FullPromt, chatItem)
					if resp.Choices != nil {
						ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content})
						ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "Ответ не верный. Подумай лучше, учти все детали задачи, и дай правильный ответ без комментариев, в своём стиле."})
						FullPromt = nil
						FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
						FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
						FullPromt = append(FullPromt, CharPrmt[gLocale]...)
						FullPromt = append(FullPromt, ChatMessages[len(ChatMessages)-4:]...)
						time.Sleep(5 * time.Second)
						resp = SendRequest(FullPromt, chatItem)
					}
				}
			case DOSHOWMENU, DOSHOWHIST, DOCLEARHIST, DOGAME:
				DoBotFunction(BotReaction, ChatMessages, update)
			case DOREADSITE:
				tmpMSGs := ProcessWebPage(LastMessages)
				FullPromt = append(FullPromt, tmpMSGs...)
				ChatMessages = append(ChatMessages, tmpMSGs...)
				resp = SendRequest(FullPromt, chatItem)
			default:
				resp = SendRequest(FullPromt, chatItem)
			}
			if BotReaction <= DOCALCULATE || BotReaction == DOREADSITE {
				msg.Text = resp.Choices[0].Message.Content
				ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: msg.Text})
				RenewDialog(update.Message.Chat.ID, ChatMessages)
				msg.Text = convTgmMarkdown(msg.Text)
				msg.ParseMode = "markdown"
				gBot.Send(msg)
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
				ProcessMember(update)
			} else {
				SendToUser(gOwner, "В группововм чате "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыли диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться в этом чате?", ACCESS, 0, update.Message.Chat.ID)
				log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
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

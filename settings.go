package main

import (
	"encoding/json"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
)

func SetChatStateDB(item ChatState) {
	var jsonData []byte
	var err error
	SetCurOperation("Set chat state", 0)
	if item.CharType < 1 {
		item.CharType = ESFJ
	}
	jsonData, err = json.Marshal(item)
	if err != nil {
		SendToUser(gOwner, gErr[11][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	} else {
		DBWrite("ChatState:"+strconv.FormatInt(item.ChatID, 10), string(jsonData), 0)
	}
}

func SetTuneChat(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Chat tuning processing", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		SendToUser(gOwner, gIm[12][gLocale], MENU_TUNE_CHAT, 1, false, chatItem.ChatID)
	}
}

func SetBotStyle(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("GPT model changing", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		chatItem.Bstyle, _ = strconv.Atoi(strings.Split(update.CallbackQuery.Data, "_ST:")[0])
		SendToUser(gOwner, "Выбран стиль общения "+gConversationStyle[chatItem.Bstyle].Name, MSG_INFO, 1, false)
		SetChatStateDB(chatItem)
	}
}

func SetBotCharacter(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Select character type", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	charValue := strings.Split(update.CallbackQuery.Data, "_")[0]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		intVal, err := strconv.Atoi(charValue)
		if err != nil {
			SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
		}
		chatItem.CharType = byte(intVal)
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "Выбран тип характера: "+gCTDescr[gLocale][chatItem.CharType-1], MSG_INFO, 1, false, chatItem.ChatID)
	}
}

func SetChatHistory(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Edit history", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		chatItem.SetState = HISTORY
		gChangeSettings = chatItem.ChatID
		SetChatStateDB(chatItem)
		if len(chatItem.History) > 0 {
			SendToUser(gOwner, "**Текущая история базовая:**\n"+gHsBasePrompt[0].Prompt[gLocale][0].Content+"\n"+
				"**Дополнитиельные факты:**\n"+chatItem.History[0].Content+"\nНапишите историю:", MSG_INFO, 1, false, chatItem.ChatID)
		} else {
			SendToUser(gOwner, "**Текущая история базовая:**\n"+gHsBasePrompt[0].Prompt[gLocale][0].Content+"\n"+
				"**Дополнитиельные факты:**\nНапишите историю:", MSG_INFO, 1, false, chatItem.ChatID)
		}
	}
}

func SetBotTemp(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Edit temperature", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		chatItem.SetState = TEMPERATURE
		gChangeSettings = chatItem.ChatID
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "Текущий уровень - "+strconv.Itoa(int(chatItem.Temperature*10))+"\nУкажите уровень экпрессии от 1 до 10", MSG_INFO, 1, false, chatItem.ChatID)
	}
}

func SetBotInitiative(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Set bot initiative", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		chatItem.SetState = INITIATIVE
		gChangeSettings = chatItem.ChatID
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "Текущий уровень - "+strconv.Itoa(int(chatItem.Inity))+"\nУкажите степень инициативы от 0 до 10", MSG_INFO, 1, false, chatItem.ChatID)
	}
}

func SetTimeZone(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Select chat time zone", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	//log.Println(chatItem.IntFacts)
	if chatItem.ChatID != 0 {
		chatItem.TimeZone, _ = strconv.Atoi(strings.Split(update.CallbackQuery.Data, "_TZ:")[0])
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "Изменено на: "+gTimezones[chatItem.TimeZone], MSG_INFO, 1, false)
		//log.Println(chatItem.IntFacts)
	}
}

func SetChatFacts(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Select chat facts", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		chatItem.InterFacts, _ = strconv.Atoi(strings.Split(update.CallbackQuery.Data, "_IF:")[0])
		SendToUser(gOwner, "Выбрана тема интересных фактов: "+gIntFacts[chatItem.InterFacts].Name, MSG_INFO, 1, false)
		SetChatStateDB(chatItem)
		SendToUser(gOwner, gIm[15][gLocale]+" "+chatIDstr, MSG_INFO, 1, false)
	}
}

func SetBotModel(update tgbotapi.Update) {
	var chatItem ChatState
	var err error
	var modelID int
	SetCurOperation("Select gpt model", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, ":")[3]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		modelID, err = strconv.Atoi(strings.Split(update.CallbackQuery.Data, ":")[1])
		if err != nil {
			SendToUser(gOwner, "Ошибка определения ID модели", MSG_ERROR, 1, false)
		}
		chatItem.Model = gModels[modelID].AI_model_name
		chatItem.AI_ID, err = strconv.Atoi(strings.Split(update.CallbackQuery.Data, ":")[2])
		if err == nil {
			SetChatStateDB(chatItem)
			SendToUser(gOwner, "Модель изменена на "+chatItem.Model+" от "+gAI[chatItem.AI_ID].AI_Name, MSG_INFO, 1, false)
		} else {
			SendToUser(gOwner, "Ошибка определения ID нейросети", MSG_ERROR, 1, false)
		}
	}
}

func SetChatSettings(chatItem ChatState, update tgbotapi.Update) {
	var temp float64
	var err error
	SetCurOperation("Set chat settings", 0)
	if chatItem.ChatID != 0 {
		switch chatItem.SetState {
		case HISTORY:
			{
				chatItem.History = []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: update.Message.Text},
					{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"}}
			}
		case TEMPERATURE:
			{
				temp, err = strconv.ParseFloat(update.Message.Text, 64)
				if err != nil {
					SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
					//log.Fatalln(err, E15[gLocale]+IM29[gLocale]+GetCurOperation())
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
					SendToUser(gOwner, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
					//log.Fatalln(err, E15[gLocale]+IM29[gLocale]+GetCurOperation())
				}
				if chatItem.Inity < 0 || chatItem.Inity > 1000 {
					chatItem.Inity = 0
				}
			}
		}
		chatItem.SetState = NO_ONE
		SetChatStateDB(chatItem)
		SendToUser(gOwner, "Принято!", MSG_INFO, 1, false)
		gChangeSettings = 0
	}
}

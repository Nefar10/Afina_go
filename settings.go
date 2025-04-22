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
	jsonData, err = json.Marshal(item)
	if err != nil {
		SendToUser(gOwner, 0, gErr[11][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
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
		SendToUser(gOwner, 0, gIm[12][gLocale], MenuTuneChat, 1, false, chatItem.ChatID)
	}
}

func SetBotStyle(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("GPT model changing", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		chatItem.CStyle, _ = strconv.Atoi(strings.Split(update.CallbackQuery.Data, "_ST:")[0])
		SendToUser(gOwner, 0, "Выбран стиль общения "+gConversationStyle[chatItem.CStyle].Name, MsgInfo, 1, false)
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
			SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
		}
		chatItem.CharType = byte(intVal)
		SetChatStateDB(chatItem)
		SendToUser(gOwner, 0, "Выбран тип характера: "+gCharTypes[chatItem.CharType-1].Description[gLocale], MsgInfo, 1, false, chatItem.ChatID)
	}
}

func SetChatHistory(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Edit history", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		chatItem.SetState = ParamHistory
		gChangeSettings = chatItem.ChatID
		SetChatStateDB(chatItem)
		if len(chatItem.History) > 0 {
			SendToUser(gOwner, 0, "**Текущая история базовая:**\n"+gHsBasePrompt[0].Prompt[gLocale][0].Content+"\n"+
				"**Дополнитиельные факты:**\n"+chatItem.History[0].Content+"\nНапишите историю:", MsgInfo, 1, false, chatItem.ChatID)
		} else {
			SendToUser(gOwner, 0, "**Текущая история базовая:**\n"+gHsBasePrompt[0].Prompt[gLocale][0].Content+"\n"+
				"**Дополнитиельные факты:**\nНапишите историю:", MsgInfo, 1, false, chatItem.ChatID)
		}
	}
}

func SetBotTemp(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Edit temperature", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		chatItem.SetState = ParamTemperature
		gChangeSettings = chatItem.ChatID
		SetChatStateDB(chatItem)
		SendToUser(gOwner, 0, "Текущий уровень - "+strconv.Itoa(int(chatItem.Temperature*10))+"\nУкажите уровень экпрессии от 1 до 10", MsgInfo, 1, false, chatItem.ChatID)
	}
}

func SetBotInitiative(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Set bot initiative", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
	if chatItem.ChatID != 0 {
		chatItem.SetState = ParamInitiative
		gChangeSettings = chatItem.ChatID
		SetChatStateDB(chatItem)
		SendToUser(gOwner, 0, "Текущий уровень - "+strconv.Itoa(int(chatItem.Inity))+"\nУкажите степень инициативы от 0 до 10", MsgInfo, 1, false, chatItem.ChatID)
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
		SendToUser(gOwner, 0, "Изменено на: "+gTimeZones[chatItem.TimeZone], MsgInfo, 1, false)
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
		SendToUser(gOwner, 0, "Выбрана тема интересных фактов: "+gIntFacts[chatItem.InterFacts].Name, MsgInfo, 1, false)
		SetChatStateDB(chatItem)
		SendToUser(gOwner, 0, gIm[15][gLocale]+" "+chatIDstr, MsgInfo, 1, false)
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
			SendToUser(gOwner, 0, "Ошибка определения ID модели", MsgError, 1, false)
		}
		chatItem.Model = gModels[modelID].AiModelName
		chatItem.AiId, err = strconv.Atoi(strings.Split(update.CallbackQuery.Data, ":")[2])
		if err == nil {
			SetChatStateDB(chatItem)
			SendToUser(gOwner, 0, "Модель изменена на "+chatItem.Model+" от "+gAI[chatItem.AiId].AiName, MsgInfo, 1, false)
		} else {
			SendToUser(gOwner, 0, "Ошибка определения ID нейросети", MsgError, 1, false)
		}
	}
}

func SetChatSettings(chatItem ChatState, update tgbotapi.Update) {
	var temp float64
	var err error
	SetCurOperation("Set chat settings", 0)
	if chatItem.ChatID != 0 {
		switch chatItem.SetState {
		case ParamHistory:
			{
				chatItem.History = []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: update.Message.Text},
					{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"}}
			}
		case ParamTemperature:
			{
				temp, err = strconv.ParseFloat(update.Message.Text, 64)
				if err != nil {
					SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
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
		case ParamInitiative:
			{
				chatItem.Inity, err = strconv.Atoi(update.Message.Text)
				if err != nil {
					SendToUser(gOwner, 0, gErr[15][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
					//log.Fatalln(err, E15[gLocale]+IM29[gLocale]+GetCurOperation())
				}
				if chatItem.Inity < 0 || chatItem.Inity > 1000 {
					chatItem.Inity = 0
				}
			}
		}
		chatItem.SetState = ParamNoOne
		SetChatStateDB(chatItem)
		SendToUser(gOwner, 0, "Принято!", MsgInfo, 1, false)
		gChangeSettings = 0
	}
}

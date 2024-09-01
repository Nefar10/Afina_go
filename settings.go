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
		chatItem.Bstyle, _ = strconv.Atoi(strings.Split(update.CallbackQuery.Data, "_ST:")[0])
		SendToUser(gOwner, "Выбран стиль общения "+gConversationStyle[chatItem.Bstyle].Name, INFO, 1)
		SetChatStateDB(chatItem)
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

func SetChatFacts(update tgbotapi.Update) {
	var chatItem ChatState
	SetCurOperation("Select chat facts")
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	//log.Println(chatItem.IntFacts)
	if chatItem.ChatID != 0 {
		chatItem.InterFacts, _ = strconv.Atoi(strings.Split(update.CallbackQuery.Data, "_IF:")[0])
		SendToUser(gOwner, "Выбрана тема интересных фактов: "+gIntFacts[chatItem.InterFacts].Name, INFO, 1)
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

func SetChatSettings(chatItem ChatState, update tgbotapi.Update) {
	var temp float64
	var err error
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
}

package main

import (
	"encoding/json"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

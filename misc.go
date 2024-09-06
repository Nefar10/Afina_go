package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
)

func SetCurOperation(msg string, log_level byte) {
	gCurProcName = msg
	if gVerboseLevel > log_level {
		log.Println(msg)
	}
}

func Log(msg string, lvl byte, err error) {
	switch lvl {
	case 0:
		log.Println(msg)
	case 1:
		log.Println(msg, err)
	case 2:
		log.Fatalln(msg, err)
	}
}

func GetChatStateDB(key string) ChatState {
	var err error
	var jsonStr string
	var chatItem ChatState
	SetCurOperation("Get chat state", 1)
	jsonStr, err = gRedisClient.Get(key).Result()
	if err != nil {
		Log("Ошибка", ERR, err)
		//SendToUser(gOwner, E13[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		return ChatState{}
	} else {
		err = json.Unmarshal([]byte(jsonStr), &chatItem)
		if err != nil {
			SendToUser(gOwner, E14[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		}
		return chatItem
	}
}

func RenewDialog(chatIDstr string, ChatMessages []openai.ChatCompletionMessage) {
	var jsonData []byte
	var err error
	SetCurOperation("Update dialog")
	jsonData, err = json.Marshal(ChatMessages)
	if err != nil {
		SendToUser(gOwner, E11[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	DBWrite("Dialog:"+chatIDstr, string(jsonData), 0)
}

func GetChatMessages(key string) []openai.ChatCompletionMessage {
	var msgString string
	var err error
	var ChatMessages []openai.ChatCompletionMessage
	SetCurOperation("Dialog reading from DB", 0)
	msgString, err = gRedisClient.Get(key).Result() //Пытаемся прочесть из БД диалог
	if err == redis.Nil {                           //Если диалога в БД нет, формируем новый и записываем в БД
		Log("Ошибка", ERR, err)
		return []openai.ChatCompletionMessage{}
	} else if err != nil {
		SendToUser(gOwner, E13[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		return []openai.ChatCompletionMessage{}
	} else { //Если диалог уже существует
		err = json.Unmarshal([]byte(msgString), &ChatMessages)
		if err != nil {
			SendToUser(gOwner, E14[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		}
		return ChatMessages
	}
}

func Restart() {
	SetCurOperation("Restarting", 0)
	SendToUser(gOwner, IM5[gLocale], INFO, 1)
	os.Exit(0)
}

func ClearContext(update tgbotapi.Update) {
	var chatItem ChatState
	var ChatMessages []openai.ChatCompletionMessage
	SetCurOperation("Context cleaning")
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

func ShowChatInfo(update tgbotapi.Update) {
	var msgString string
	var chatItem ChatState
	SetCurOperation("Chat info view", 0)
	chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
	chatItem = GetChatStateDB("ChatState:" + chatIDstr)
	if chatItem.ChatID != 0 {
		msgString = "Название чата: " + chatItem.Title + "\n" +
			"Модель поведения: " + gConversationStyle[chatItem.Bstyle].Name + "\n" +
			"Тип характера: " + gCTDescr[gLocale][chatItem.CharType-1] + "\n" +
			"Нейронная сеть: " + chatItem.Model + "\n" +
			"Экспрессия: " + strconv.FormatFloat(float64(chatItem.Temperature*100), 'f', -1, 32) + "%\n" +
			"Инициативность: " + strconv.Itoa(chatItem.Inity*10) + "%\n" +
			"Тема интересных фактов: " + gIntFacts[chatItem.InterFacts].Name + "\n" +
			"Текущая версия: " + VER
		SendToUser(chatItem.ChatID, msgString, INFO, 2)
	}
}

func CheckChatRights(update tgbotapi.Update) {
	var err error
	var jsonStr string       //Current json string
	var questItem QuestState //Current QuestState item
	var ansItem Answer       //Curent Answer intem
	var chatItem ChatState
	SetCurOperation("Chat state changing")
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
}

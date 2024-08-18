package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/sashabaranov/go-openai"
)

func SetCurOperation(msg string) {
	gCurProcName = msg
	if gVerboseLevel > 0 {
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
	jsonStr, err = gRedisClient.Get(key).Result()
	if err == redis.Nil {
		SendToUser(gOwner, E16[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, INFO, 0)
		return ChatState{}
	} else if err != nil {
		SendToUser(gOwner, E13[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		return ChatState{}
	} else {
		err = json.Unmarshal([]byte(jsonStr), &chatItem)
		if err != nil {
			SendToUser(gOwner, E14[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
		}
		return chatItem
	}
}

func SetChatStateDB(item ChatState) {
	var jsonData []byte
	var err error
	//Checks
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

func RenewDialog(chatIDstr string, ChatMessages []openai.ChatCompletionMessage) {
	var jsonData []byte
	var err error
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

func DBWrite(key string, value string, expiration time.Duration) error {
	var err = gRedisClient.Set(key, value, expiration).Err()
	if err != nil {
		SendToUser(gOwner, E10[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	return err
}

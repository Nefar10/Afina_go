package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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
		//SendToUser(gOwner, E13[gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
		return ChatState{AllowState: IN_PROCESS}
	} else {
		err = json.Unmarshal([]byte(jsonStr), &chatItem)
		if err != nil {
			SendToUser(gOwner, gErr[14][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
		}
		return chatItem
	}
}

func RenewDialog(chatID int64, ChatMessages []openai.ChatCompletionMessage) {
	var jsonData []byte
	var err error
	var chatIDstr string
	chatIDstr = strconv.FormatInt(chatID, 10)
	SetCurOperation("Update dialog", 0)
	jsonData, err = json.Marshal(ChatMessages)
	if err != nil {
		SendToUser(gOwner, gErr[11][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
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
		SendToUser(gOwner, gErr[13][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
		return []openai.ChatCompletionMessage{}
	} else { //Если диалог уже существует
		err = json.Unmarshal([]byte(msgString), &ChatMessages)
		if err != nil {
			SendToUser(gOwner, gErr[14][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
		}
		return ChatMessages
	}
}

func Restart() {
	SetCurOperation("Restarting", 0)
	SendToUser(gOwner, gIm[5][gLocale], INFO, 1)
	os.Exit(0)
}

func ClearContext(chatID int64) {
	var ChatMessages []openai.ChatCompletionMessage
	SetCurOperation("Context cleaning", 0)
	ChatMessages = nil
	RenewDialog(chatID, ChatMessages)
	SendToUser(chatID, "Контекст очищен!", NOTHING, 1)
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
	SetCurOperation("Chat state changing", 0)
	err = json.Unmarshal([]byte(update.CallbackQuery.Data), &ansItem)
	if err == nil {
		jsonStr, err = gRedisClient.Get("QuestState:" + ansItem.CallbackID.String()).Result() //читаем состояние запрса из БД
		if err == redis.Nil {
			SendToUser(gOwner, gErr[16][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, INFO, 0)
		} else if err != nil {
			SendToUser(gOwner, gErr[13][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
		} else {
			err = json.Unmarshal([]byte(jsonStr), &questItem)
			if err != nil {
				SendToUser(gOwner, gErr[14][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
			}
			if questItem.State == QUEST_IN_PROGRESS {
				chatItem = GetChatStateDB("ChatState:" + strconv.FormatInt(questItem.ChatID, 10))
				if chatItem.ChatID != 0 {
					switch ansItem.State { //Изменяем флаг доступа
					case ALLOW:
						{
							chatItem.AllowState = ALLOW
							SendToUser(gOwner, gIm[6][gLocale], INFO, 0)
							SendToUser(chatItem.ChatID, gIm[7][gLocale], INFO, 1)
						}
					case DISALLOW:
						{
							chatItem.AllowState = DISALLOW
							SendToUser(gOwner, gIm[8][gLocale], INFO, 0)
							SendToUser(chatItem.ChatID, gIm[9][gLocale], INFO, 1)
						}
					case BLACKLISTED:
						{
							chatItem.AllowState = BLACKLISTED
							SendToUser(gOwner, gIm[10][gLocale], INFO, 0)
							SendToUser(chatItem.ChatID, gIm[11][gLocale], INFO, 1)
						}
					}
					SetChatStateDB(chatItem)
				}
			}
		}
	}
}

func isNow(update tgbotapi.Update, timezone int) [][]openai.ChatCompletionMessage {
	var lHsTime [][]openai.ChatCompletionMessage
	currentTime := time.Unix(int64(update.Message.Date+((timezone-15)*3600)), 0)
	timeString := currentTime.Format("Monday, 2006-01-02 15:04:05")
	lHsTime = append(lHsTime, []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "Current time is " + timeString + "."}})
	lHsTime = append(lHsTime, []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "Текущее время " + timeString + "."}})
	return lHsTime
}

func convTgmMarkdown(input string) string {
	clean := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x9F]+`)
	input = clean.ReplaceAllString(input, "")
	boldPattern := regexp.MustCompile(`\*\*(.*?)\*\*`)
	input = boldPattern.ReplaceAllString(input, "*$1*")
	boldPattern2 := regexp.MustCompile(`__(.*?)__`)
	input = boldPattern2.ReplaceAllString(input, "*$1*")
	/*
		italicPattern := regexp.MustCompile(`\*(.*?)\*`)
		input = italicPattern.ReplaceAllString(input, "=$1=")
		boldPattern3 := regexp.MustCompile(`==(.*?)==`)
		input = boldPattern3.ReplaceAllString(input, "*$1*")
		italicPattern2 := regexp.MustCompile(`=(.*?)=`)
		input = italicPattern2.ReplaceAllString(input, "_ $1 _")
	*/
	return input
}

func sendHistory(chatID int64, ChatMessages []openai.ChatCompletionMessage) {
	var buffer bytes.Buffer

	for _, msg := range ChatMessages {
		_, err := fmt.Fprintf(&buffer, "%s: %s\n", msg.Role, msg.Content)
		if err != nil {
			SendToUser(gOwner, "Ошибка формирования сообщения", ERROR, 2)
			return
		}
	}

	msg := tgbotapi.NewDocument(chatID, tgbotapi.FileReader{
		Name:   "Messages.txt", // Имя файла, которое будет отображаться в Telegram
		Reader: &buffer,
	})

	if _, err := gBot.Send(msg); err != nil {
		SendToUser(gOwner, "Ошибка при отправке документа", ERROR, 2)
		return
	}
}

func SendRequest(FullPrompt []openai.ChatCompletionMessage, chatItem ChatState) openai.ChatCompletionResponse {
	var resp openai.ChatCompletionResponse
	var err error
	// Формируем запрос к API
	gClient_is_busy = true    //Флаг занятости
	gLastRequest = time.Now() //Запомним текущее время
	resp, err = gClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       chatItem.Model,
			Temperature: chatItem.Temperature,
			Messages:    FullPrompt,
		},
	)
	if err != nil {
		SendToUser(gOwner, gErr[17][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, INFO, 0)
		time.Sleep(20 * time.Second)
	}
	gClient_is_busy = false
	return resp
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
)

// SetCurOperation Storing current operation for logging and debugging
func SetCurOperation(msg string, logLevel byte) {
	if len(msg) > 0 {
		gSysMutex.Lock()
		gCurProcName = msg
		gSysMutex.Unlock()
		if gVerboseLevel > logLevel {
			Log(msg, ErrNo, nil)
		}
	} else {
		return
	}
}

// GetCurOperation Reading current operation for logging and debugging
func GetCurOperation() string {
	gSysMutex.Lock()
	curOp := gCurProcName
	gSysMutex.Unlock()
	return curOp
}

// Log Events logging
func Log(msg string, lvl byte, err error) {
	if len(msg) > 0 {
		switch lvl {
		case 0:
			log.Println(msg)
		case 1:
			log.Println(msg, err)
		case 2:
			log.Fatalln(msg, err)
		}
	} else {
		return
	}
}

// ParseChatKeyID Converts redis string chatID to telegram int64
func ParseChatKeyID(key string) int64 {
	var s string
	var n int64
	var err error
	SetCurOperation("Determining chat ID", 1)
	if len(key) <= 0 {
		return 0
	}
	if strings.Contains("ChatState:", s) {
		s = strings.Split(key, ":")[1]
	} else {
		if gVerboseLevel > 1 {
			SendToUser(gOwner, 0, "Ошибка парсинга при извлечении ID чата из строки "+key, MsgError, 2, false)
		} else {
			Log("Ошибка парсинга при извлечении ID чата из строки "+key, Err, nil)
		}
		return 0
	}
	n, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		if gVerboseLevel > 1 {
			SendToUser(gOwner, 0, "Ошибка парсинга ID чата при преобразовании в int "+s, MsgError, 2, false)
		} else {
			Log("Ошибка парсинга ID чата при преобразовании в int "+s, Err, err)
		}
		return 0
	} else {
		return n
	}
}

func GetChatStateDB(chatID int64) ChatState {
	var err error
	var jsonStr string
	var chatItem ChatState
	SetCurOperation("Get chat state", 1)
	jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(chatID, 10)).Result()
	if err != nil {
		Log("Ошибка", Err, err)
		if chatID == gOwner {
			return gDefChatState
		} else {
			return ChatState{AllowState: ChatInProcess, ChatID: 0}
		}
	} else {
		err = json.Unmarshal([]byte(jsonStr), &chatItem)
		if err != nil {
			SendToUser(gOwner, 0, gErr[14][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
			return ChatState{AllowState: ChatInProcess, ChatID: 0}
		} else {
			return chatItem
		}
	}
}

func UpdateDialog(chatID int64, ChatMessages []openai.ChatCompletionMessage) {
	var chatIDstr string
	SetCurOperation("Update dialog", 0)
	chatIDstr = strconv.FormatInt(chatID, 10)
	jsonData, err := json.Marshal(ChatMessages)
	if err != nil {
		SendToUser(gOwner, 0, gErr[11][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
		return
	} else {
		DBWrite("Dialog:"+chatIDstr, string(jsonData), 0)
	}
}

func GetDialog(key string) []openai.ChatCompletionMessage {
	var msgString string
	var err error
	var ChatMessages []openai.ChatCompletionMessage
	SetCurOperation("Dialog reading from DB", 0)
	msgString, err = gRedisClient.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			Log("Не найдена запись в БД", Err, err)
		} else {
			SendToUser(gOwner, 0, gErr[13][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
		}
		return []openai.ChatCompletionMessage{}
	} else {
		err = json.Unmarshal([]byte(msgString), &ChatMessages)
		if err != nil {
			SendToUser(gOwner, 0, gErr[14][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
			return []openai.ChatCompletionMessage{}
		} else {
			return ChatMessages
		}
	}
}

func Restart() {
	SetCurOperation("Restarting", 0)
	SendToUser(gOwner, 0, gIm[5][gLocale], MsgInfo, 1, false)
	os.Exit(0)
}

func ClearContext(chatID int64) {
	var ChatMessages []openai.ChatCompletionMessage
	SetCurOperation("Context cleaning", 0)
	ChatMessages = nil
	UpdateDialog(chatID, ChatMessages)
	SendToUser(chatID, 0, "Контекст очищен!", MsgNothing, 1, false)
}

func ShowChatInfo(update tgbotapi.Update) {
	var msgString string
	var chatItem ChatState
	var chatIDstr string
	SetCurOperation("Chat info view", 0)
	chatIDstr = strings.Split(update.CallbackQuery.Data, " ")[1]
	if len(chatIDstr) <= 0 {
		Log("Ошибка парсинга ID чата", Err, nil)
		return
	} else {
		chatItem = GetChatStateDB(ParseChatKeyID("ChatState:" + chatIDstr))
		if chatItem.ChatID != 0 {
			msgString = "Название чата: " + chatItem.Title + "\n" +
				"Модель поведения: " + gConversationStyle[chatItem.CStyle].Name + "\n" +
				"Тип характера: " + gCharTypes[chatItem.CharType-1].Description[gLocale] + "\n" +
				"Нейронная сеть: " + chatItem.Model + " от " + gAI[chatItem.AiId].AiName + "\n" +
				"Модель принятия решений: " + gAI[chatItem.AiId].AiBaseModel + "\n" +
				"Экспрессия: " + strconv.FormatFloat(float64(chatItem.Temperature*100), 'f', -1, 32) + "%\n" +
				"Инициативность: " + strconv.Itoa(chatItem.Inity*10) + "%\n" +
				"Тема интересных фактов: " + gIntFacts[chatItem.InterFacts].Name + "\n" +
				"Текущая версия: " + Ver + "\n" +
				"ID чата: " + chatIDstr
			SendToUser(chatItem.ChatID, 0, msgString, MsgInfo, 2, false)
		}
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
		jsonStr, err = gRedisClient.Get("QuestState:" + ansItem.CallbackID.String()).Result()
		if err != nil {
			SendToUser(gOwner, 0, gErr[13][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
			if err == redis.Nil {
				SendToUser(gOwner, 0, gErr[16][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgInfo, 0, false)
			}
			return
		} else {
			err = json.Unmarshal([]byte(jsonStr), &questItem)
			if err != nil {
				SendToUser(gOwner, 0, gErr[14][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
				return
			}
			if questItem.State == QuestInProgress {
				chatItem = GetChatStateDB(questItem.ChatID)
				if chatItem.ChatID != 0 {
					switch ansItem.State {
					case ChatAllow:
						{
							chatItem.AllowState = ChatAllow
							SendToUser(gOwner, 0, gIm[6][gLocale], MsgInfo, 0, false)
							SendToUser(chatItem.ChatID, 0, gIm[7][gLocale], MsgInfo, 1, false)
						}
					case ChatDisallow:
						{
							chatItem.AllowState = ChatDisallow
							SendToUser(gOwner, 0, gIm[8][gLocale], MsgInfo, 0, false)
							SendToUser(chatItem.ChatID, 0, gIm[9][gLocale], MsgInfo, 1, false)
						}
					case ChatBlacklist:
						{
							chatItem.AllowState = ChatBlacklist
							SendToUser(gOwner, 0, gIm[10][gLocale], MsgInfo, 0, false)
							SendToUser(chatItem.ChatID, 0, gIm[11][gLocale], MsgInfo, 1, false)
						}
					}
					SetChatStateDB(chatItem)
				}
			}
		}
	}
}

func isNow(currentTime time.Time) [][]openai.ChatCompletionMessage {
	SetCurOperation("Determining current time", 0)
	var lHsTime [][]openai.ChatCompletionMessage
	timeString := currentTime.Format("Monday, 2006-01-02 15:04:05")
	lHsTime = append(lHsTime, []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "Current time is " + timeString + "."}})
	lHsTime = append(lHsTime, []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "Текущее время " + timeString + "."}})
	return lHsTime
}

func convTgmMarkdown(input string) string {
	var clean, bdPat *regexp.Regexp
	var err error
	SetCurOperation("Fomatting message", 1)
	if len(input) <= 0 {
		Log("Сообщение отсутсвует", Err, nil)
		return ""
	} else {
		clean, err = regexp.Compile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x9F]+`)
		if err == nil {
			input = clean.ReplaceAllString(input, "")
		}
		//itPat, err = regexp.Compile(`(\n|\s)\*([^*].+?)\*`)
		//if err == nil {
		//	input = itPat.ReplaceAllString(input, "$1\u200B_\u200B$2\u200B_")
		//}
		bdPat, err = regexp.Compile(`\*\*(.+?)\*\*`)
		if err == nil {
			input = bdPat.ReplaceAllString(input, "*$1*")
		}
		bdPat, err = regexp.Compile(`#(#*?)(\s.+?)\n`)
		if err == nil {
			input = bdPat.ReplaceAllString(input, "`$2`\n")
		}
		return input
	}
}

func sendHistory(chatID int64, ChatMessages []openai.ChatCompletionMessage) {
	var buffer bytes.Buffer
	SetCurOperation("Processing history", 0)
	if len(ChatMessages) > 0 {
		for _, msg := range ChatMessages {
			_, err := fmt.Fprintf(&buffer, "%s: %s\n", msg.Role, msg.Content)
			if err != nil {
				SendToUser(gOwner, 0, err.Error()+" Ошибка буферизации", MsgError, 2, false)
				return
			}
		}
		msg := tgbotapi.NewDocument(chatID, tgbotapi.FileReader{
			Name:   "Messages.txt",
			Reader: &buffer,
		})
		if _, err := gBot.Send(msg); err != nil {
			Log("Ошибка при отправке документа", Err, err)
			return
		}
	}
}

func SendRequest(FullPrompt []openai.ChatCompletionMessage, chatItem ChatState) string {
	var resp openai.ChatCompletionResponse
	var err error
	//var YaFullprompt []yandexgpt.YandexGPTMessage
	//var prmt openai.ChatCompletionMessage
	//log.Println(FullPrompt)
	SetCurOperation("Request to AI", 0)
	/*
		if chatItem.ChatID == gOwner {
			SetCurOperation("Request to Yandex AI", 0)
			for _, prmt = range FullPrompt {
				YaFullprompt = append(YaFullprompt, yandexgpt.YandexGPTMessage{Role: yandexgpt.YandexGPTMessageRoleSystem, Text: prmt.Content})
			}
			request := yandexgpt.YandexGPTRequest{
				ModelURI: yandexgpt.MakeModelURI("b1g9hte57kfq967nevga", yandexgpt.YandexGPT4Model),
				CompletionOptions: yandexgpt.YandexGPTCompletionOptions{
					Stream:      false,
					Temperature: chatItem.Temperature,
					MaxTokens:   2000,
				},
				Messages: YaFullprompt,
			}

			response, err := gYaClient.GetCompletion(context.Background(), request)
			if err != nil {
				fmt.Println("Request error", err.Error())
				return ""
			}
			return response.Result.Alternatives[0].Message.Text
		}
	*/
	gAIMutex.Lock()
	gLastRequest = time.Now() //Запомним текущее время
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	resp, err = gClient[chatItem.AiId].CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       chatItem.Model, //"deepseek-chat", //"deepseek/deepseek-r1:free",
			Temperature: chatItem.Temperature,
			Messages:    FullPrompt,
		},
	)
	gAIMutex.Unlock()
	if err != nil {
		SendToUser(gOwner, 0, gErr[17][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgInfo, 0, false)
		return ""
	}
	return resp.Choices[0].Message.Content
}

func BotWaiting(ChatID int64, tm int) {
	SetCurOperation("BotWaiting", 1)
	act := tgbotapi.NewChatAction(ChatID, tgbotapi.ChatTyping)
	gBot.Send(act)
	for {
		currentTime := time.Now()
		elapsedTime := currentTime.Sub(gLastRequest)
		time.Sleep(time.Second)
		if elapsedTime >= 3*time.Second {
			break
		}
	}
	gAIMutex.Lock()
	gLastRequest = time.Now()
	gAIMutex.Unlock()
}

func saveCustomPrompts(filename string, prompts []CustomPrompt) error {
	data, err := json.MarshalIndent(prompts, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadCustomPrompts(filename string) ([]CustomPrompt, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var prompts []CustomPrompt
	if err := json.Unmarshal(data, &prompts); err != nil {
		return nil, err
	}

	return prompts, nil
}

func saveMsgs(filename string, msgs [][2]string) error {
	data, err := json.MarshalIndent(msgs, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadMsgs(filename string) ([][2]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var msgs [][2]string
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, err
	}

	return msgs, nil
}

func NewTask(id int64, chatid int64, description, prompt string, isRecurring bool, interval time.Duration, priority int) Task {
	now := time.Now()
	return Task{
		ID:                id,
		Description:       description,
		Prompt:            prompt,
		CreatedAt:         now,
		LastExecutedAt:    now,
		ChatID:            chatid,
		IsRecurring:       isRecurring,
		Interval:          interval,
		NextExecutionTime: now.Add(interval), // Устанавливаем следующее время выполнения, исходя из интервала
		Priority:          priority,
		State:             TaskNew, // Начальный статус
	}
}

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func GetFileURL(fileID string, update tgbotapi.Update) string {
	var file tgbotapi.File
	var fileURL string
	var err error
	file, err = gBot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		if update.Message.Chat.Type == "private" {
			SendToUser(update.Message.Chat.ID, 0, "Не удалось получить идентификатор файла "+err.Error(), MsgError, 1, true)
		}
		return ""
	}

	fileURL, err = gBot.GetFileDirectURL(file.FileID)
	if err != nil {
		if update.Message.Chat.Type == "private" {
			SendToUser(update.Message.Chat.ID, 0, "Не удалось получить прямую ссылку для загрузки "+err.Error(), MsgError, 1, true)
		}
		return ""
	}
	return fileURL
}

package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChatState struct { //Структура для хранения настроек чатов
	ChatID     int64  //Идентификатор чата
	UserName   string //Имя пользователя
	AllowState int    //Флаг разрешения/запрещения доступа
	BotState   int    //Состояние бота в чате
}

type QuestState struct { //струдктура для оперативного хранения вопросов
	ChatID   int64 //идентификатор чатов
	Question int   //тип запроса
	State    int   //состояние обработки
}

var gBot *tgbotapi.BotAPI      //Указатель на бота
var gToken string              //API токен бота
var gOwner int64               //Владелец бота сюда будут приходить служебные сообщения и вопросы от бота
var gBotName string            //Имя бота, на которое бот будет отзываться в групповом чате
var gBotGender int             //Пол бота оказывает влияние на его представление
var gChatsStates []ChatState   //Для инициализации списка доступов для чатов. Сохраняется в файл
var gQuestsStates []QuestState //Для слежения за квестами в реальном времени
var gRedisIP string            //Адрес сервера БД
var gRedisPASS string          //Пароль к redis
var gRedisClient *redis.Client //Клиент redis
var gDir string                //Для хранения текущей директории

func SendToOwner(mesText string, quest int) { //отправка сообщения владельцу
	msg := tgbotapi.NewMessage(gOwner, mesText) //инициализируем сообщение
	switch quest {                              //разбираем, вдруг требуется отправить запрос
	case ACCESS: //В случае, если стоит вопрос доступа
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Да", "ACCEPTED"),
					tgbotapi.NewInlineKeyboardButtonData("Нет", "DISACCEPTED"),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case MENU: //В случае, если стоит вопрос доступа
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Разрешенные собеседники", "WHITELIST"),
					tgbotapi.NewInlineKeyboardButtonData("Запрещенные собеседники", "BLACKLIST"),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	}
	gBot.Send(msg) //отправляем сообщение
}

func init() {
	var err error
	var owner int
	var jsonData []byte
	var jsonString string
	var chatState ChatState
	gRedisIP = os.Getenv(REDIS_IN_OS)                       //Читаем адрес сервера БД из переменных окружения
	gRedisPASS = os.Getenv(REDIS_PASS_IN_OS)                //Читаем пароль к серверу БД из переенных окружения
	gToken = os.Getenv(TOKEN_NAME_IN_OS)                    //читаем токен бота из переменных окружения
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil { //инициализируем бота
		log.Panic(err)
	} else {
		gBot.Debug = true
	}
	if owner, err = strconv.Atoi(os.Getenv(OWNER_IN_OS)); err != nil { //читаем идентификатор владельца из переменных окружения
		log.Panic(err)
	} else {
		gOwner = int64(owner) //запоминаем идентификатор
	}
	gBotName = os.Getenv(BOTNAME_IN_OS) //читаем и запоминаем имя ота из переменных окружения
	switch os.Getenv(BOTGENDER_IN_OS) { //читаем пол бота из переменных окружения
	case "Male":
		gBotGender = MALE
	case "Female":
		gBotGender = FEMALE
	default:
		gBotGender = NEUTRAL
	}
	if gDir, err = os.Getwd(); err != nil {
		log.Panic("Ошибка при определении текущей директории:", err)
	}
	gRedisClient = redis.NewClient(&redis.Options{ // Создаем клиент Redis
		Addr:     gRedisIP,   // адрес Redis сервера
		Password: gRedisPASS, // пароль, если необходимо
		DB:       0,          // номер базы данных
	})
	_, err = gRedisClient.Ping().Result() // Проверяем соединение
	if err != nil {
		SendToOwner("Не удается установить соединение с СУБД по причине\n'"+err.Error()+"'", NOTHING)
		log.Panic(err)
		return
	}
	keys, err := gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		panic(err)
	}
	if len(keys) > 0 {
		for _, key := range keys {
			jsonString, err = gRedisClient.Get(key).Result()
			if err != nil {
				panic(err)
			}
			err := json.Unmarshal([]byte(jsonString), &chatState)
			if err != nil {
				panic(err)
			}
			gChatsStates = append(gChatsStates, chatState)
		}
		log.Println(gChatsStates)
	} else {
		gChatsStates = []ChatState{{ChatID: 0, AllowState: DISALLOW, UserName: "All", BotState: SLEEP},
			{ChatID: gOwner, AllowState: ALLOW, UserName: "Owner", BotState: RUN}} //инициализируем массив для сохранения состояний чатов

		for _, item := range gChatsStates {
			jsonData, err = json.Marshal(item)
			err = gRedisClient.Set("ChatState:"+strconv.FormatInt(item.ChatID, 10), string(jsonData), 0).Err()
			if err != nil {
				log.Panic(err)
			}
			SendToOwner("Иницализирую ключи настройки чатов", NOTHING)
		}
	}
	/*
			err = gRedisClient.Set("ChatState:"+strconv.FormatInt(gChatsStates[0].ChatID, 10), string(jsonData), 0).Err()
			if err != nil {
				panic(err)
			}
			jsonData, err = json.Marshal(gChatsStates[1])
			err = gRedisClient.Set("ChatState:"+strconv.FormatInt(gChatsStates[1].ChatID, 10), string(jsonData), 0).Err()
			if err != nil {
				panic(err)
			}
		}

		/*
			if !fileexists(gDir + FILES_ALLOW_LIST) { //проверям существование файла со списокм чатов
					log.Printf("File %s not exists", gDir+FILES_ALLOW_LIST)
					file, err := os.OpenFile(gDir+FILES_ALLOW_LIST, os.O_WRONLY|os.O_CREATE, 0644)
					if err != nil {
						log.Fatal(err)
					}
					defer file.Close()
					writer := bufio.NewWriter(file)
					// Записываем каждую строку в файл
					for _, curChat := range gChatsStates {
						_, err = writer.WriteString(strconv.FormatInt(curChat.ChatID, 10) + " " + curChat.UserName + " " + strconv.Itoa(curChat.AllowState) + " " + strconv.Itoa(curChat.BotState) + "\n")
						if err != nil {
							log.Fatal(err)
						}
					}
					err = writer.Flush()
					if err != nil {
						log.Fatal(err)
					}
					SendToOwner("Сведения о состоянии чатов были инициализированы", NOTHING)
					defer file.Close()
				} else {
					file, err := os.Open(gDir + FILES_ALLOW_LIST)
					if err != nil {
						log.Fatal(err)
					}
					defer file.Close()
					scanner := bufio.NewScanner(file)
					for scanner.Scan() {
						line := scanner.Text()
						elements := strings.Split(line, " ") // Разделяем строку на отдельные значения, используя пробел в качестве разделителя
						// Создаем новую структуру и заполняем значениями
						var data ChatState
						data.ChatID, err = strconv.ParseInt(elements[0], 10, 64)
						if err != nil {
							log.Fatal(err)
						}
						data.UserName = elements[1]
						data.AllowState, err = strconv.Atoi(elements[2])
						if err != nil {
							log.Fatal(err)
						}
						data.BotState, err = strconv.Atoi(elements[3])
						if err != nil {
							log.Fatal(err)
						}
						gChatsStates = append(gChatsStates, data)
					}

					if err := scanner.Err(); err != nil {
						log.Fatal(err)
					}
				}*/
	SendToOwner("Я снова на связи", NOTHING) //отправляем приветственное сообщение владельцу
}

func main() {
	log.Printf("Authorized on account %s", gBot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if update.Message.IsCommand() {
				command := update.Message.Command()
				switch command {
				case "menu":
					if update.Message.Chat.ID == gOwner {
						SendToOwner("Выберите, что необходимо сделать", MENU)
					} else {
						SendToOwner("Выберите, что необходимо сделать", USERMENU)
					}
				}
			}
			//AllowChat(update.Message.Chat.ID, update.Message.From.FirstName+" "+update.Message.From.UserName, update.Message.Text)
			//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			//msg.ReplyToMessageID = update.Message.MessageID

			//gBot.Send(msg)
		}
	}
}

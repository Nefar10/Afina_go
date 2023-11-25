package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type ChatState struct { //Структура для хранения настроек чатов
	ChatID     int64  //Идентификатор чата
	UserName   string //Имя пользователя
	AllowState int    //Флаг разрешения/запрещения доступа
	BotState   int    //Состояние бота в чате
}

type QuestState struct { //струдктура для оперативного хранения вопросов
	ChatID   int64     //идентификатор чатов
	Question int       //тип запроса
	State    int       //состояние обработки
	Time     time.Time //текущее время
}

var gBot *tgbotapi.BotAPI      //Указатель на бота
var gToken string              //API токен бота
var gOwner int64               //Владелец бота сюда будут приходить служебные сообщения и вопросы от бота
var gBotName string            //Имя бота, на которое бот будет отзываться в групповом чате
var gBotGender int             //Пол бота оказывает влияние на его представление
var gChatsStates []ChatState   //Для инициализации списка доступов для чатов. Сохраняется в файл
var gQuestsStates []QuestState //Для слежения за квестами в реальном времени
var gRedisIP string            //Адрес сервера БД
var gRedisDB int               //Используемая БД 0-15
var gRedisPASS string          //Пароль к redis
var gRedisClient *redis.Client //Клиент redis
var gDir string                //Для хранения текущей директории

func SendToOwner(mesText string, quest int, chatID ...int64) { //отправка сообщения владельцу
	var jsonData []byte
	var err error
	var item QuestState
	msg := tgbotapi.NewMessage(gOwner, mesText) //инициализируем сообщение
	switch quest {                              //разбираем, вдруг требуется отправить запрос
	case ACCESS: //В случае, если стоит вопрос доступа
		{
			callbackID := uuid.New()
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Да", "ACCEPTED:"+callbackID.String()),
					tgbotapi.NewInlineKeyboardButtonData("Нет", "DISACCEPTED:"+callbackID.String()),
					tgbotapi.NewInlineKeyboardButtonData("Блокировать", "BLАCKLISTED:"+callbackID.String()),
				))
			msg.ReplyMarkup = numericKeyboard
			item.ChatID = chatID[0]
			item.Question = quest
			item.State = IN_PROCESS
			item.Time = time.Now()
			jsonData, err = json.Marshal(item)
			if err != nil {
				log.Panic(err)
			}
			err = gRedisClient.Set("QuestState:"+callbackID.String(), string(jsonData), 0).Err()
			if err != nil {
				log.Panic(err)
			}
		}
	case MENU: //Вызвано меню администратора
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
	var db int
	var jsonData []byte
	var jsonString string
	var chatState ChatState
	gRedisIP = os.Getenv(REDIS_IN_OS)                                 //Читаем адрес сервера БД из переменных окружения
	gRedisPASS = os.Getenv(REDIS_PASS_IN_OS)                          //Читаем пароль к серверу БД из переенных окружения
	if db, err = strconv.Atoi(os.Getenv(REDISDB_IN_OS)); err != nil { //читаем идентификатор БД из переменных окружения
		log.Panic(err)
	} else {
		gRedisDB = db //запоминаем идентификатор БД
	}
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
		DB:       gRedisDB,   // номер базы данных
	})
	_, err = gRedisClient.Ping().Result() // Проверяем соединение
	if err != nil {
		SendToOwner("Не удается установить соединение с СУБД по причине\n'"+err.Error()+"'", NOTHING)
		log.Panic(err)
		return
	}
	keys, err := gRedisClient.Keys("ChatState:*").Result() //Пытаемся прочесть все ключи чатов
	if err != nil {
		log.Panic(err)
	}
	if len(keys) > 0 { //Если ключи были считаны - запомнить их
		for _, key := range keys {
			jsonString, err = gRedisClient.Get(key).Result()
			if err != nil {
				log.Panic(err)
			}
			err = json.Unmarshal([]byte(jsonString), &chatState)
			if err != nil {
				log.Panic(err)
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
	SendToOwner("Я снова на связи", NOTHING) //отправляем приветственное сообщение владельцу
}

func main() {
	var err error
	var item ChatState
	var itemStr string
	log.Printf("Authorized on account %s", gBot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if update.Message.IsCommand() { //Начало обработки команд
				command := update.Message.Command()
				switch command {
				case "menu":
					if update.Message.Chat.ID == gOwner {
						SendToOwner("Выберите, что необходимо сделать", MENU)
					} else {
						SendToOwner("Выберите, что необходимо сделать", USERMENU)
					}
				}
			} else { //Начало обработки сообщений
				itemStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(update.Message.Chat.ID, 10)).Result()
				if err == redis.Nil {
					log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName)
					AllowChat(update.Message.Chat.ID, update.Message.From.FirstName+" "+update.Message.From.UserName, update.Message.Text)
				} else if err != nil {
					log.Fatal(err)
				} else {
					err = json.Unmarshal([]byte(itemStr), &item)
					switch item.AllowState {
					case ALLOW:
						{
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Принято")
							msg.ReplyToMessageID = update.Message.MessageID
							gBot.Send(msg)
							log.Println(itemStr)
							log.Println(item)
							//Здесь начинается обработка сообщения
						}
					case DISALLOW:
						{
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Запрещено")
							msg.ReplyToMessageID = update.Message.MessageID
							gBot.Send(msg)
						}
					case BLACKLISTED:
						{
							log.Panic("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName)
						}
					}
				}
			}
		}
	}
}

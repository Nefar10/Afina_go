package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

type ChatState struct { //Структура для хранения настроек чатов
	ChatID     int64  //Идентификатор чата
	UserName   string //Имя пользователя
	AllowState int    //Флаг разрешения/запрещения доступа
	BotState   int    //Состояние бота в чате
}

type QuestState struct { //струдктура для оперативного хранения вопросов
	ChatID     int64     //идентификатор чатов
	CallbackID uuid.UUID //идентификатор запроса
	Question   int       //тип запроса
	State      int       //состояние обработки
	Time       time.Time //текущее время
}

type Answer struct { //Структура callback
	CallbackID uuid.UUID //идентификатор вопроса
	State      int       //ответ
}

var gBot *tgbotapi.BotAPI      //Указатель на бота
var gToken string              //API токен бота
var gOwner int64               //Владелец бота сюда будут приходить служебные сообщения и вопросы от бота
var gBotName string            //Имя бота, на которое бот будет отзываться в групповом чате
var gBotGender int             //Пол бота оказывает влияние на его представление
var gChatsStates []ChatState   //Для инициализации списка доступов для чатов. Сохраняется в файл
var gRedisIP string            //Адрес сервера БД
var gRedisDB int               //Используемая БД 0-15
var gAIToken string            //AI API ключ
var gRedisPASS string          //Пароль к redis
var gRedisClient *redis.Client //Клиент redis
var gDir string                //Для хранения текущей директории
var gLastRequest time.Time

func SendToOwner(mesText string, quest int, chatID ...int64) { //отправка сообщения владельцу
	var jsonData []byte
	var jsonDataAllow []byte
	var jsonDataDeny []byte
	var jsonDataBlock []byte
	var err error
	var item QuestState
	var ans Answer
	msg := tgbotapi.NewMessage(gOwner, mesText) //инициализируем сообщение
	switch quest {                              //разбираем, вдруг требуется отправить запрос
	case ACCESS: //В случае, если стоит вопрос доступа формируем меню запроса
		{
			callbackID := uuid.New()                                                             //создаем уникальный идентификатор запроса
			item.ChatID = chatID[0]                                                              //указываем ID чата источника
			item.Question = quest                                                                //указывам тип запроса
			item.CallbackID = callbackID                                                         //запоминаем уникальнй ID
			item.State = QUEST_IN_PROGRESS                                                       //соотояние обработки, которое запишем в БД
			item.Time = time.Now()                                                               //запомним текущее время
			jsonData, _ = json.Marshal(item)                                                     //конвертируем структуру в json
			err = gRedisClient.Set("QuestState:"+callbackID.String(), string(jsonData), 0).Err() //Делаем запись в БД
			if err != nil {
				log.Panic(err)
			}
			ans.CallbackID = item.CallbackID
			ans.State = ALLOW //Генерируем вариант ответа "разрешить" для callback
			jsonDataAllow, _ = json.Marshal(ans)
			ans.State = DISALLOW //генерируем вариант ответа "запретить" для callback
			jsonDataDeny, _ = json.Marshal(ans)
			ans.State = BLACKLISTED //генерируем вариант ответа "заблокаировать" для callback
			jsonDataBlock, _ = json.Marshal(ans)
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Да", string(jsonDataAllow)),
					tgbotapi.NewInlineKeyboardButtonData("Нет", string(jsonDataDeny)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Блокировать", string(jsonDataBlock)),
				))
			msg.ReplyMarkup = numericKeyboard

		}
	case MENU: //Вызвано меню администратора
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Разрешенные чаты", "WHITELIST"),
					tgbotapi.NewInlineKeyboardButtonData("Запрещенные чаты", "BLACKLIST"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Полный сброс", "RESETTODEFAULTS"),
					tgbotapi.NewInlineKeyboardButtonData("Очистка кеша", "FLUSHCACHE"),
					tgbotapi.NewInlineKeyboardButtonData("Перезагрузка", "RESTART"),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case USERMENU:
		{

		}
	}
	gBot.Send(msg) //отправляем сообщение
}

func init() {
	var err error
	var owner int
	var db int
	var jsonData []byte
	gRedisIP = os.Getenv(REDIS_IN_OS)                                 //Читаем адрес сервера БД из переменных окружения
	gRedisPASS = os.Getenv(REDIS_PASS_IN_OS)                          //Читаем пароль к серверу БД из переенных окружения
	if db, err = strconv.Atoi(os.Getenv(REDISDB_IN_OS)); err != nil { //читаем идентификатор БД из переменных окружения
		log.Panic(err)
	} else {
		gRedisDB = db //запоминаем идентификатор БД
	}
	gToken = os.Getenv(TOKEN_NAME_IN_OS)                    //читаем токен бота из переменных окружения
	gAIToken = os.Getenv(AI_IN_OS)                          //Читаем токен OpenAI из переменных окружения
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil { //инициализируем бота
		log.Panic(err)
	} else {
		//gBot.Debug = true
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
	gChatsStates = []ChatState{{ChatID: 0, AllowState: DISALLOW, UserName: "All", BotState: SLEEP},
		{ChatID: gOwner, AllowState: ALLOW, UserName: "Owner", BotState: RUN}} //инициализируем массив для сохранения состояний чатов

	for _, item := range gChatsStates {
		jsonData, err = json.Marshal(item)
		err = gRedisClient.Set("ChatState:"+strconv.FormatInt(item.ChatID, 10), string(jsonData), 0).Err()
		if err != nil {
			log.Panic(err)
		}
	}
	SendToOwner("Я снова на связи", NOTHING) //отправляем приветственное сообщение владельцу
}

func main() {
	var err error            //Для реакции на ошибки
	var itemStr string       //Оперативное хранение строки json
	var jsonData []byte      //Строка json конвертированная в byte-код
	var chatItem ChatState   //Для оперативного хранения структуры ChatState
	var questItem QuestState //Для оперативного хранения структуры QuestState
	var ansItem Answer       //Для оперативного хранения структуры Answer
	var keys []string        //Для оперативного хранения считанных ключей
	var msgString string     //Для формирования сообщения
	var prompt string        //Для формирования promt для AI
	log.Printf("Authorized on account %s", gBot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)
	client := openai.NewClient(gAIToken)

	for update := range updates {
		if update.CallbackQuery != nil { //Если прилетел ответ на вопрос
			switch update.CallbackQuery.Data {
			case "WHITELIST":
				{
					keys, err = gRedisClient.Keys("ChatState:*").Result() //Пытаемся прочесть все ключи чатов
					if err != nil {
						log.Panic(err)
					}
					if len(keys) > 0 { //Если ключи были считаны - запомнить их
						msgString = "Список разрешенных чатов\n"
						for _, key := range keys {
							itemStr, err = gRedisClient.Get(key).Result()
							if err != nil {
								log.Panic(err)
							}
							err = json.Unmarshal([]byte(itemStr), &chatItem)
							if err != nil {
								log.Panic(err)
							}
							if chatItem.AllowState == ALLOW {
								msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " - " + chatItem.UserName + "\n"
							}
						}
						msg := tgbotapi.NewMessage(gOwner, msgString)
						gBot.Send(msg)
					}

				}
			case "BLACKLIST":
				{
					keys, err = gRedisClient.Keys("ChatState:*").Result() //Пытаемся прочесть все ключи чатов
					if err != nil {
						log.Panic(err)
					}
					if len(keys) > 0 { //Если ключи были считаны - запомнить их
						msgString = "Список запрещенных чатов\n"
						for _, key := range keys {
							itemStr, err = gRedisClient.Get(key).Result()
							if err != nil {
								log.Panic(err)
							}
							err = json.Unmarshal([]byte(itemStr), &chatItem)
							if err != nil {
								log.Panic(err)
							}
							if chatItem.AllowState != ALLOW {
								msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " - " + chatItem.UserName + "\n"
							}
						}
						msg := tgbotapi.NewMessage(gOwner, msgString)
						gBot.Send(msg)
					}

				}
			case "RESETTODEFAULTS":
				{
					keys, err = gRedisClient.Keys("ChatState:*").Result() //Пытаемся прочесть все ключи чатов
					if err != nil {
						log.Panic(err)
					}
					if len(keys) > 0 { //Если ключи были считаны - запомнить их
						msgString = "Информация о доступах для следующих чатов была очищена\n"
						for _, key := range keys {
							itemStr, err = gRedisClient.Get(key).Result()
							if err != nil {
								log.Panic(err)
							}
							err = json.Unmarshal([]byte(itemStr), &chatItem)
							if err != nil {
								log.Panic(err)
							}
							if chatItem.ChatID != 0 && chatItem.ChatID != gOwner {
								msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " - " + chatItem.UserName + "\n"
								err = gRedisClient.Del("ChatState:" + strconv.FormatInt(chatItem.ChatID, 10)).Err()
								if err != nil {
									log.Fatal(err)
								}
							}
						}
						msg := tgbotapi.NewMessage(gOwner, msgString)
						gBot.Send(msg)
					}
				}
			case "FLUSHCACHE":
				{
					keys, err = gRedisClient.Keys("QuestState:*").Result() //Пытаемся прочесть все ключи чатов
					if err != nil {
						log.Panic(err)
					}
					if len(keys) > 0 { //Если ключи были считаны - запомнить их
						msgString = "Кеш очищен\n"
						for _, key := range keys {
							err = gRedisClient.Del(key).Err()
							if err != nil {
								log.Panic(err)
							}
						}
						msgString = msgString + "Было удалено " + strconv.Itoa(len(keys)) + " устаревших записей"
						msg := tgbotapi.NewMessage(gOwner, msgString)
						gBot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(gOwner, "Очищать нечего")
						gBot.Send(msg)
					}
				}
			case "RESTART":
				{
					msg := tgbotapi.NewMessage(gOwner, "Перезагружаюсь")
					gBot.Send(msg)
					os.Exit(0)
				}
			default:
				{
					err = json.Unmarshal([]byte(update.CallbackQuery.Data), &ansItem)                     //Разбираем поступивший ответ
					itemStr, err = gRedisClient.Get("QuestState:" + ansItem.CallbackID.String()).Result() //читаем состояние запрса из БД
					if err == redis.Nil {
						log.Println("Запись БД " + "QuestState:" + ansItem.CallbackID.String() + " не айдена")
					} else if err != nil {
						log.Panic(err)
					} else {
						err = json.Unmarshal([]byte(itemStr), &questItem) //Разбираем считанные из БД данные
						if questItem.State == QUEST_IN_PROGRESS {         //Если к нам прилетело решение запроса доступа - разбираем его
							itemStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(questItem.ChatID, 10)).Result() //Читаем состояние чата
							if err == redis.Nil {
								log.Println("Запись БД ChatState:" + strconv.FormatInt(questItem.ChatID, 10) + " не айдена")
							} else if err != nil {
								log.Fatal(err)
							} else {
								err = json.Unmarshal([]byte(itemStr), &chatItem) //Разбираем считанные из БД данные
							}
							switch ansItem.State { //Изменяем флаг доступа
							case ALLOW:
								{
									chatItem.AllowState = ALLOW
									SendToOwner("Доступ предоставлен", NOTHING)
									msg := tgbotapi.NewMessage(chatItem.ChatID, "Мне позволили с Вами общаться!")
									gBot.Send(msg)
								}
							case DISALLOW:
								{
									chatItem.AllowState = DISALLOW
									SendToOwner("Доступ запрещен", NOTHING)
									msg := tgbotapi.NewMessage(chatItem.ChatID, "Прошу прощения, для продолжения общения необхоимо оформить подписку.")
									gBot.Send(msg)
								}
							case BLACKLISTED:
								{
									chatItem.AllowState = BLACKLISTED
									SendToOwner("Доступ заблокирован", NOTHING)
									msg := tgbotapi.NewMessage(chatItem.ChatID, "Поздравляю! Вы были добавлены в список проказников!")
									gBot.Send(msg)
								}
							}
							jsonData, err = json.Marshal(chatItem) //Конвертируем новое состояние чата в json и записываем в тот же ключ БД
							err = gRedisClient.Set("ChatState:"+strconv.FormatInt(questItem.ChatID, 10), string(jsonData), 0).Err()
							if err != nil {
								log.Panic(err)
							}
						}
					}
				}
			}

			continue
		}
		if update.Message != nil { // If we got a message
			//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
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
				continue
			}
			if update.Message.Text != "" { //Начало обработки простого сообщения
				itemStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(update.Message.Chat.ID, 10)).Result()
				if err == redis.Nil {
					log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName)
					chatItem.ChatID = update.Message.Chat.ID
					chatItem.BotState = RUN
					chatItem.AllowState = IN_PROCESS
					chatItem.UserName = update.Message.From.UserName
					jsonData, err = json.Marshal(chatItem)
					err = gRedisClient.Set("ChatState:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err()
					if err != nil {
						log.Panic(err)
					}
					SendToOwner("Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, update.Message.Chat.ID)
				} else if err != nil {
					log.Fatal(err)
				} else {
					err = json.Unmarshal([]byte(itemStr), &chatItem)
					if chatItem.BotState == RUN {
						switch chatItem.AllowState {
						case ALLOW:
							{
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
								msg.ReplyToMessageID = update.Message.MessageID
								prompt, err = gRedisClient.Get("Dialog:" + strconv.FormatInt(update.Message.Chat.ID, 10)).Result()
								if err == redis.Nil {
									prompt = update.Message.From.FirstName + ": " + update.Message.Text + "\n"
									err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), prompt, 0).Err()
									if err != nil {
										log.Panic(err)
									}
								} else if err != nil {
									log.Panic(err)
								} else {
									prompt = prompt + update.Message.From.FirstName + ": " + update.Message.Text + "\n"
									err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), prompt, 0).Err()
									if err != nil {
										log.Panic(err)
									}
								}
								action := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
								gBot.Send(action)
								for {
									currentTime := time.Now()
									elapsedTime := currentTime.Sub(gLastRequest)

									if elapsedTime >= 20*time.Second {
										break
									}
								}
								gLastRequest = time.Now()
								gBot.Send(action)
								resp, err := client.CreateChatCompletion(
									context.Background(),
									openai.ChatCompletionRequest{
										Model: openai.GPT3Dot5Turbo1106,
										Messages: []openai.ChatCompletionMessage{
											{
												Role:    openai.ChatMessageRoleUser,
												Content: prompt,
											},
										},
									},
								)

								if err != nil {
									log.Printf("ChatCompletion error: %v\n", err)
									return
								}
								msg.Text = resp.Choices[0].Message.Content
								prompt = prompt + msg.Text
								err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), prompt, 0).Err()
								if err != nil {
									log.Panic(err)
								}
								gBot.Send(msg)
							}
						case DISALLOW:
							{
								SendToOwner("Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, update.Message.Chat.ID)

								//gBot.Send(msg)
							}
						case BLACKLISTED:
							{
								log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName)
							}
						}
					}
				}
			}
		}
	}
}

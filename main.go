package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

type ChatState struct { //Структура для хранения настроек чатов
	ChatID      int64   //Идентификатор чата
	UserName    string  //Имя пользователя
	AllowState  int     //Флаг разрешения/запрещения доступа
	BotState    int     //Состояние бота в чате
	Type        string  //Тип чата private,group,supergroup
	Model       string  //Выбранная для чата модель общения
	Temperature float32 //Креативность бота
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
var gBotNames []string         //Имя бота, на которое бот будет отзываться в групповом чате
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
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Играть в IT-Элиас", "GAME_IT_ALIAS"),
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
	gBotNames = strings.Split(os.Getenv(BOTNAME_IN_OS), ",") //читаем и запоминаем имя бота из переменных окружения
	switch os.Getenv(BOTGENDER_IN_OS) {                      //читаем пол бота из переменных окружения
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
	var err error                                   //Для реакции на ошибки
	var itemStr string                              //Оперативное хранение строки json
	var jsonData []byte                             //Строка json конвертированная в byte-код
	var chatItem ChatState                          //Для оперативного хранения структуры ChatState
	var questItem QuestState                        //Для оперативного хранения структуры QuestState
	var ansItem Answer                              //Для оперативного хранения структуры Answer
	var keys []string                               //Для оперативного хранения считанных ключей
	var msgString string                            //Для формирования сообщения
	var prompt string                               //Для формирования promt для AI
	var ChatMessages []openai.ChatCompletionMessage //Для оработки диалога
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
				itemStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(update.Message.Chat.ID, 10)).Result() //Читаем инфо от чате в БД
				if err == redis.Nil {                                                                                  //Если записи в БД нет - формирруем новую запись
					log.Println("Запрос нового диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName)
					chatItem.ChatID = update.Message.Chat.ID
					chatItem.BotState = RUN
					chatItem.AllowState = IN_PROCESS //Указываем статус допуска
					chatItem.UserName = update.Message.From.UserName
					chatItem.Type = update.Message.Chat.Type
					chatItem.Model = openai.GPT3Dot5Turbo1106
					chatItem.Temperature = 0.9
					jsonData, err = json.Marshal(chatItem)
					err = gRedisClient.Set("ChatState:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err() //Записываем инфо о чате в БД
					if err != nil {
						log.Panic(err)
					}
					SendToOwner("Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, update.Message.Chat.ID)
				} else if err != nil { //Тут может вылезти какая-нибудь ошибкаа доступа к БД
					log.Panicln(err)
				} else { //Если мы успешно считали информацию о чате в БД, то переодим к рвоерке прав
					err = json.Unmarshal([]byte(itemStr), &chatItem)
					if chatItem.BotState == RUN {
						switch chatItem.AllowState { //Если доступ предоставлен
						case ALLOW:
							{
								ChatMessages = nil                                     //Формируем новый диалог
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "") //Формирум новый ответ
								if update.Message.Chat.Type != "private" {             //Если чат не приватный, то ставим отметку - на какое соощение отвечаем
									msg.ReplyToMessageID = update.Message.MessageID
								}
								msgString, err = gRedisClient.Get("Dialog:" + strconv.FormatInt(update.Message.Chat.ID, 10)).Result() //Пытаемся прочесть из БД диалог
								if err == redis.Nil {                                                                                 //Если диалога в БД нет, формируем новый и записываем в БД
									//TO DO Добавить дефолтный prompt в начало диалога
									if update.Message.Chat.Type == "private" { //Если текущий чат приватный
										ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.Text})
										//prompt = update.Message.Text + "\n" //Записываем первое сообщение чата
									} else { //Если текущи чат групповой записываем первое сообщение чата дополняя его именем текущего собеседника
										ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.From.FirstName + ": " + update.Message.Text})
										//prompt = update.Message.From.FirstName + ": " + update.Message.Text + "\n"
									}
									jsonData, err = json.Marshal(ChatMessages)
									err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err() //Записываем диалог в БД
									if err != nil {                                                                                            //Здесь могут быть всякие ошибки записи в БД
										log.Panic(err)
									}
								} else if err != nil { //Здесь могут быть всякие ошибки чтения из БД
									log.Panic(err)
								} else { //Если диалог уже существует
									err = json.Unmarshal([]byte(msgString), &ChatMessages)
									if update.Message.Chat.Type == "private" { //Если текущий чат приватный
										ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.Text})
										//prompt = prompt + update.Message.Text + "\n"
									} else { //Если текущи чат групповой дописываем сообщение чата дополняя его именем текущего собеседника
										ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.From.FirstName + ": " + update.Message.Text})
										//prompt = prompt + update.Message.From.FirstName + ": " + update.Message.Text + "\n"
									}
									jsonData, err = json.Marshal(ChatMessages)
									err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err() //Записываем диалог в БД
									if err != nil {                                                                                            //Здесь могут быть всякие ошибки записи в БД
										log.Panic(err)
									}
								}
								action := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
								//gBot.Send(action) //Здесь мы делаем вид, что бот отреагировал на новое сообщение
								for { //Здесь мы делаем паузу, позволяющую не отправлять промпты чаще чем раз в 20 секунд
									currentTime := time.Now()
									elapsedTime := currentTime.Sub(gLastRequest)

									if elapsedTime >= 20*time.Second {
										break
									}
								}
								switch update.Message.Chat.Type { //Здесь мы обрабатываем запросы к openAI для различных чатов
								case "private":
									{
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
									}
								default: //Если у нас е приватный чат, что ведем себя как в группе
									{
										for _, name := range gBotNames { //Определим - есть ли в контексте последнего сообщения имя бота
											if (strings.Contains(update.Message.Text, name)) || (update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == gBot.Self.ID) { //Если имя бота встречается
												gLastRequest = time.Now() //Прежде чем формировать запрос, запомним текущее время
												for i := 0; i < 3; i++ {
													gBot.Send(action)                         //Здесь мы продолжаем делать вид, что бот отреагировал на новое сообщение
													resp, err := client.CreateChatCompletion( //Формируем запрос к мозгам
														context.Background(),
														openai.ChatCompletionRequest{
															Model:       chatItem.Model,
															Temperature: chatItem.Temperature,
															MaxTokens:   4000,
															Messages:    ChatMessages,
														},
													)
													if err != nil {
														log.Printf("ChatCompletion error: %v\n", err)
														log.Panicln("Предпримем попытку еще одного запроса")
													} else {
														log.Printf("Токенов использовано %d", resp.Usage.TotalTokens)
														msg.Text = resp.Choices[0].Message.Content //Записываем ответ в сообщение
														break
													}
												}
												break
											}
										}

									}
								}
								ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: msg.Text})
								//prompt = prompt + msg.Text
								jsonData, err = json.Marshal(ChatMessages)
								err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err()
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

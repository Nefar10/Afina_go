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

func SendToUser(toChat int64, mesText string, quest int, chatID ...int64) { //отправка сообщения владельцу
	var jsonData []byte                         //Для оперативного хранения
	var jsonDataAllow []byte                    //Для формирования uuid ответа ДА
	var jsonDataDeny []byte                     //Для формирования uuid ответа НЕТ
	var jsonDataBlock []byte                    //Для формирования uuid ответа Блок
	var err error                               //Временное хранение ошибок
	var item QuestState                         //Для хранения состояния колбэка
	var ans Answer                              //Для формирования uuid колбэка
	msg := tgbotapi.NewMessage(toChat, mesText) //инициализируем сообщение
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
			if err != nil {                                                                      //Тут могут быть ошибки записи в БД
				log.Panic(err)
			}
			ans.CallbackID = item.CallbackID //Генерируем вариант ответа "разрешить" для callback
			ans.State = ALLOW
			jsonDataAllow, _ = json.Marshal(ans) //генерируем вариант ответа "запретить" для callback
			ans.State = DISALLOW
			jsonDataDeny, _ = json.Marshal(ans) //генерируем вариант ответа "заблокаировать" для callback
			ans.State = BLACKLISTED
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
	case USERMENU: //Меню подписчика
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Изменить параметры", "TUNE_CHAT"),
					tgbotapi.NewInlineKeyboardButtonData("История чата", "CHAT_PROMPT"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Играть в IT-Элиас", "GAME_IT_ALIAS: "+strconv.FormatInt(toChat, 10)),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case SELECTCHAT:
		{
			msg.Text = "Выерите чат для настройки"
			chats := strings.Split(mesText, "\n")
			var buttons []tgbotapi.InlineKeyboardButton
			for _, chat := range chats {
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(chat, strings.Split(mesText, "~")[0]))
			}
			var rows [][]tgbotapi.InlineKeyboardButton
			for _, button := range buttons {
				row := []tgbotapi.InlineKeyboardButton{button}
				rows = append(rows, row)
			}
			numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.ReplyMarkup = numericKeyboard
		}
	case TUNECHAT: //меню настройки чата
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Выбрать модель", "GPT_MODEL"),
					tgbotapi.NewInlineKeyboardButtonData("Креативность", "MODEL_TEMP"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Размер контекста", "CONTEXT_LEN"),
					tgbotapi.NewInlineKeyboardButtonData("История чата", "CHAT_PROMPT"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Права доступа", "RIGHTS"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", "MENU"),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	}
	gBot.Send(msg) //отправляем сообщение
}

func init() {
	var err error //Временное хранение ошибок
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
		SendToUser(gOwner, "Не удается установить соединение с СУБД по причине\n'"+err.Error()+"'", NOTHING)
		log.Panic(err)
		return
	}
	//инициализируем массив для сохранения состояний чатов
	gChatsStates = append(gChatsStates, ChatState{ChatID: 0, Model: openai.GPT3Dot5Turbo1106, Temperature: 1, AllowState: DISALLOW, UserName: "All", BotState: SLEEP, Type: "private", History: gHsNulled})
	gChatsStates = append(gChatsStates, ChatState{ChatID: gOwner, Model: openai.GPT3Dot5Turbo1106, Temperature: 1, AllowState: ALLOW, UserName: "Owner", BotState: RUN, Type: "private", History: gHsOwner})

	for _, item := range gChatsStates {
		jsonData, err = json.Marshal(item)
		err = gRedisClient.Set("ChatState:"+strconv.FormatInt(item.ChatID, 10), string(jsonData), 0).Err()
		if err != nil {
			log.Panic(err)
		}
	}
	SendToUser(gOwner, "Я снова на связи", NOTHING) //отправляем приветственное сообщение владельцу
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
	var ChatMessages []openai.ChatCompletionMessage //Для оработки диалога
	log.Printf("Authorized on account %s", gBot.Self.UserName)
	updateConfig := tgbotapi.NewUpdate(0) //Инициализируем канал получения обновлений
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)
	client := openai.NewClient(gAIToken) //Инициализируем клиента
	for update := range updates {        //Обрабатываем сообщения с канала обновлений
		if update.CallbackQuery != nil { //Если прилетел ответ на вопрос
			if update.CallbackQuery.Data == "WHITELIST" || update.CallbackQuery.Data == "BLACKLIST" {
				keys, err = gRedisClient.Keys("ChatState:*").Result() //Пытаемся прочесть все ключи чатов
				if err != nil {
					log.Panic(err)
				}
				if len(keys) > 0 { //Если ключи были считаны - запомнить их

					for _, key := range keys {
						itemStr, err = gRedisClient.Get(key).Result()
						if err != nil {
							log.Panic(err)
						}
						err = json.Unmarshal([]byte(itemStr), &chatItem)
						if err != nil {
							log.Panic(err)
						}
						if chatItem.AllowState == ALLOW && update.CallbackQuery.Data == "WHITELIST" {
							if chatItem.Type != "private" {
								msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
							} else {
								msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
							}
						}
						if chatItem.AllowState != ALLOW && update.CallbackQuery.Data == "BLACKLIST" {
							if chatItem.Type != "private" {
								msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
							} else {
								msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
							}
						}
					}
					SendToUser(gOwner, msgString, SELECTCHAT)
				}
			}
			if update.CallbackQuery.Data == "RESETTODEFAULTS" {
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
							if chatItem.Type != "private" {
								msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " - " + chatItem.Type + "\n"
							} else {
								msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " - " + chatItem.UserName + "\n"
							}
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
			if update.CallbackQuery.Data == "FLUSHCACHE" {
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
			if update.CallbackQuery.Data == "RESTART" {
				msg := tgbotapi.NewMessage(gOwner, "Перезагружаюсь")
				gBot.Send(msg)
				os.Exit(0)
			}
			if strings.Contains(update.CallbackQuery.Data, "ID:") {
				chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
				chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
				if err != nil {
					log.Panicln(err)
				}
				SendToUser(update.CallbackQuery.From.ID, "Выберите действие c чатом "+chatIDstr, TUNECHAT, chatID)
			}
			if update.CallbackQuery.Data == "CHAT_PROMPT" {
				itemStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(update.CallbackQuery.From.ID, 10)).Result() //Читаем инфо от чате в БД
				if err == redis.Nil {                                                                                        //Если записи в БД нет - формирруем новую запись
					msg := tgbotapi.NewMessage(chatItem.ChatID, "Предыстория отсутствует")
					gBot.Send(msg)
				} else if err != nil { //Тут может вылезти какая-нибудь ошибкаа доступа к БД
					log.Panicln(err)
				} else {
					err = json.Unmarshal([]byte(itemStr), &chatItem) //Выведем существующий prompt сообщением
					jsonData, err = json.Marshal(chatItem.History)
					msg := tgbotapi.NewMessage(chatItem.ChatID, string(jsonData))
					msg.Text = msg.Text + "\n\nДавай-ка сотворим историю!"
					gBot.Send(msg)
				}
			}
			if strings.Contains(update.CallbackQuery.Data, "GAME_IT_ALIAS") {
				chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
				chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
				if err != nil {
					log.Panicln(err)
				}
				itemStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result() //Читаем инфо от чате в БД
				if err == redis.Nil {
					log.Panicln(err)
				} else if err != nil {
					log.Panicln(err)
				} else {
					err = json.Unmarshal([]byte(msgString), &ChatMessages)
					ChatMessages = append(ChatMessages, gITAlias...)
					jsonData, err = json.Marshal(ChatMessages)
					err = gRedisClient.Set("Dialog:"+chatIDstr, string(jsonData), 0).Err() //Записываем диалог в БД
					if err != nil {                                                        //Здесь могут быть всякие ошибки записи в БД
						log.Panic(err)
					}
					SendToUser(chatID, "Пишите - как только будете готовы начать игру.", NOTHING)
				}
			}
			if update.CallbackQuery.Data == "MENU" {
				if update.CallbackQuery.From.ID == gOwner {
					SendToUser(gOwner, "Выберите, что необходимо сделать", MENU)
				} else {
					SendToUser(update.CallbackQuery.From.ID, "Выберите, что необходимо сделать", USERMENU)
				}
			}
			err = json.Unmarshal([]byte(update.CallbackQuery.Data), &ansItem)
			if err == nil {
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
								SendToUser(gOwner, "Доступ предоставлен", NOTHING)
								msg := tgbotapi.NewMessage(chatItem.ChatID, "Мне позволили с Вами общаться!")
								gBot.Send(msg)
							}
						case DISALLOW:
							{
								chatItem.AllowState = DISALLOW
								SendToUser(gOwner, "Доступ запрещен", NOTHING)
								msg := tgbotapi.NewMessage(chatItem.ChatID, "Прошу прощения, для продолжения общения необхоимо оформить подписку.")
								gBot.Send(msg)
							}
						case BLACKLISTED:
							{
								chatItem.AllowState = BLACKLISTED
								SendToUser(gOwner, "Доступ заблокирован", NOTHING)
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
			continue
		}
		if update.Message != nil { // If we got a message
			//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if update.Message.IsCommand() { //Начало обработки команд
				command := update.Message.Command()
				switch command {
				case "menu":
					if update.Message.Chat.ID == gOwner {
						SendToUser(gOwner, "Выберите, что необходимо сделать", MENU)
					} else {
						SendToUser(update.Message.Chat.ID, "Выберите, что необходимо сделать", USERMENU)
					}
				}
				continue
			}
			if update.Message.Text != "" { //Начало обработки простого сообщения
				itemStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(update.Message.Chat.ID, 10)).Result() //Читаем инфо от чате в БД
				if err == redis.Nil {                                                                                  //Если записи в БД нет - формирруем новую запись
					chatItem.ChatID = update.Message.Chat.ID
					chatItem.BotState = RUN
					chatItem.AllowState = IN_PROCESS //Указываем статус допуска
					chatItem.UserName = update.Message.From.UserName
					chatItem.Type = update.Message.Chat.Type
					chatItem.Title = update.Message.Chat.Title
					chatItem.Model = openai.GPT3Dot5Turbo1106
					chatItem.Temperature = 1.1
					chatItem.History = gHsOwner
					jsonData, err = json.Marshal(chatItem)
					err = gRedisClient.Set("ChatState:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err() //Записываем инфо о чате в БД
					if err != nil {
						log.Panic(err)
					}
					if update.Message.Chat.Type == "private" {
						log.Println("Запрос нового диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName)
						SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, update.Message.Chat.ID)
					} else {
						log.Println("Запрос нового диалога от группового чата " + update.Message.From.FirstName + " " + update.Message.Chat.Title)
						SendToUser(gOwner, "В группововм чате "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыли диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, update.Message.Chat.ID)
					}
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
									ChatMessages = append(ChatMessages, chatItem.History...)
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
										gLastRequest = time.Now() //Прежде чем формировать запрос, запомним текущее время
										for i := 0; i < 3; i++ {
											gBot.Send(action)                         //Здесь мы продолжаем делать вид, что бот отреагировал на новое сообщение
											resp, err := client.CreateChatCompletion( //Формируем запрос к мозгам
												context.Background(),
												openai.ChatCompletionRequest{
													Model:       chatItem.Model,
													Temperature: chatItem.Temperature,
													Messages:    ChatMessages,
												},
											)
											if err != nil {
												log.Printf("ChatCompletion error: %v\n", err)
												log.Panicln("Предпримем попытку еще одного запроса")
											} else {
												log.Printf("Чат ID: %d Токенов использовано: %d", update.Message.Chat.ID, resp.Usage.TotalTokens)
												msg.Text = resp.Choices[0].Message.Content //Записываем ответ в сообщение
												break
											}
										}
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
															Messages:    ChatMessages,
														},
													)
													if err != nil {
														log.Printf("ChatCompletion error: %v\n", err)
														log.Panicln("Предпримем попытку еще одного запроса")
													} else {
														log.Printf("Чат ID: %d Токенов использовано: %d", update.Message.Chat.ID, resp.Usage.TotalTokens)
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
								if update.Message.Chat.Type == "private" {
									SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, update.Message.Chat.ID)
								} else {
									SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, update.Message.Chat.ID)

								}

								//gBot.Send(msg)
							}
						case BLACKLISTED:
							{
								if update.Message.Chat.Type == "private" {
									log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName)
								} else {
									log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title)
								}
							}
						case IN_PROCESS:
							{
								if update.Message.Chat.Type == "private" {
									log.Println("Запрос нового диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName)
									SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nВопрос не был решен ранее", NOTHING)
								} else {
									log.Println("Запрос нового диалога от группового чата " + update.Message.From.FirstName + " " + update.Message.Chat.Title)
									SendToUser(gOwner, "В группововм чате "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыли диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nВопрос не был решен ранее", NOTHING)
								}
							}
						}
					}
				}
			}
		}
	}
}

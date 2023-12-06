package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

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
	case ERROR:
		{
			msg.Text = mesText + "\n" + IM0
		}
	case INFO:
		{
			msg.Text = mesText
		}
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
				log.Fatalln(err)
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
					tgbotapi.NewInlineKeyboardButtonData("Без решения", "INPROCESS"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Полный сброс", "RESETTODEFAULTS"),
					tgbotapi.NewInlineKeyboardButtonData("Очистка кеша", "FLUSHCACHE"),
					tgbotapi.NewInlineKeyboardButtonData("Перезагрузка", "RESTART"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Изменить параметры", "TUNE_CHAT: "+strconv.FormatInt(toChat, 10)),
					tgbotapi.NewInlineKeyboardButtonData("Очистить контекст", "CLEAR_CONTEXT: "+strconv.FormatInt(toChat, 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Играть в IT-Элиас", "GAME_IT_ALIAS: "+strconv.FormatInt(toChat, 10)),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case USERMENU: //Меню подписчика
		{
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup( //формируем меню для ответа
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Изменить параметры", "TUNE_CHAT: "+strconv.FormatInt(toChat, 10)),
					tgbotapi.NewInlineKeyboardButtonData("Очистить контекст", "CLEAR_CONTEXT: "+strconv.FormatInt(toChat, 10)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Играть в IT-Элиас", "GAME_IT_ALIAS: "+strconv.FormatInt(toChat, 10)),
				))
			msg.ReplyMarkup = numericKeyboard
		}
	case SELECTCHAT:
		{
			msg.Text = "Выберите чат для настройки"
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
	//Temporary variables
	var err error
	var owner int
	var db int
	var jsonData []byte
	gRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	//Read bot API key from OS env
	gToken = os.Getenv(TOKEN_NAME_IN_OS)
	if gToken == "" {
		log.Fatalln(E1 + TOKEN_NAME_IN_OS)
	}
	//Read owner's chatID from OS env
	owner, err = strconv.Atoi(os.Getenv(OWNER_IN_OS))
	if err != nil {
		log.Fatalln(err, E2+OWNER_IN_OS)
	} else {
		gOwner = int64(owner) //Storing owner's chat ID in variable
	}
	//Telegram bot init
	gBot, err = tgbotapi.NewBotAPI(gToken)
	if err != nil {
		log.Fatalln(err, E6)
	} else {
		//gBot.Debug = true
		log.Printf("Authorized on account %s", gBot.Self.UserName)
	}
	//Current dir init
	gDir, err = os.Getwd()
	if err != nil {
		SendToUser(gOwner, E8+err.Error(), ERROR)
		log.Fatalln(err, E8)
	}
	//Read redis connector options from OS env
	//Redis IP
	gRedisIP = os.Getenv(REDIS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, E3+REDIS_IN_OS, ERROR)
		log.Fatalln(E3 + REDIS_IN_OS)
	}
	//Redis password
	gRedisPass = os.Getenv(REDIS_PASS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, E4+REDIS_PASS_IN_OS, ERROR)
		log.Fatalln(E4 + REDIS_PASS_IN_OS)
	}
	//DB ID
	db, err = strconv.Atoi(os.Getenv(REDISDB_IN_OS))
	if err != nil {
		SendToUser(gOwner, E5+REDISDB_IN_OS+err.Error(), ERROR)
		log.Fatalln(E5 + REDIS_PASS_IN_OS)
	} else {
		gRedisDB = db //Storing DB ID
	}
	//Redis client init
	gRedisClient = redis.NewClient(&redis.Options{
		Addr:     gRedisIP,
		Password: gRedisPass,
		DB:       gRedisDB,
	})
	//Chek redis connection
	err = redisPing(*gRedisClient)
	if err != nil {
		SendToUser(gOwner, E9+err.Error(), ERROR)
		log.Fatalln(err, E9)
	}
	//Read OpenAI API token from OS env
	gAIToken = os.Getenv(AI_IN_OS)
	if gAIToken == "" {
		SendToUser(gOwner, E7+AI_IN_OS, ERROR)
		log.Fatalln(E7 + AI_IN_OS)
	}
	//Read bot names from OS env
	gBotNames = strings.Split(strings.ToUpper(os.Getenv(BOTNAME_IN_OS)), ",")
	if gBotNames[0] == "" {
		SendToUser(gOwner, IM1+BOTNAME_IN_OS, INFO)
		gBotNames = []string{"AFINA", "АФИНА"}
		log.Println(IM1 + BOTNAME_IN_OS)
	}
	// Default prompt init
	var gHsOwner = []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль универсального персонального асисстента. Зови себя - " + gBotNames[0] + "."},
		{Role: openai.ChatMessageRoleAssistant, Content: "Здравствуйте. Понял, можете называть меня " + gBotNames[0] + ". Я Ваш универсальный ассистент."}}
	//Read bot gender from OS env
	switch os.Getenv(BOTGENDER_IN_OS) {
	case "Male":
		gBotGender = MALE
	case "Female":
		gBotGender = FEMALE
	default:
		SendToUser(gOwner, IM2+BOTGENDER_IN_OS, INFO)
		gBotGender = NEUTRAL
		log.Println(IM2 + BOTGENDER_IN_OS)
	}
	//Default chat states init
	gChatsStates = append(gChatsStates, ChatState{ChatID: 0, Model: openai.GPT3Dot5Turbo1106, Inity: 0, Temperature: 1, AllowState: DISALLOW, UserName: "All", BotState: SLEEP, Type: "private", History: gHsNulled})
	gChatsStates = append(gChatsStates, ChatState{ChatID: gOwner, Model: openai.GPT3Dot5Turbo1106, Inity: 5, Temperature: 1, AllowState: ALLOW, UserName: "Owner", BotState: RUN, Type: "private", History: gHsOwner})
	//Storing default chat states to DB
	for _, item := range gChatsStates {
		jsonData, err = json.Marshal(item)
		if err != nil {
			SendToUser(gOwner, E11+err.Error(), ERROR)
			log.Fatalln(err)
		} else {
			err = gRedisClient.Set("ChatState:"+strconv.FormatInt(item.ChatID, 10), string(jsonData), 0).Err()
			if err != nil {
				SendToUser(gOwner, E10+err.Error(), ERROR)
				log.Fatalln(err, E10)
			}
		}
	}
	//Send init complete message to owner
	SendToUser(gOwner, IM3, INFO)
	log.Println("Initialization complete!")

}

func main() {
	//Temporary variables
	var err error                                   //Some errors
	var jsonStr string                              //Current json string
	var jsonData []byte                             //Current json bytecode
	var chatItem ChatState                          //Current ChatState item
	var questItem QuestState                        //Current QuestState item
	var ansItem Answer                              //Curent Answer intem
	var keys []string                               //Curent keys array
	var msgString string                            //Current message string
	var ChatMessages []openai.ChatCompletionMessage //Current prompt

	//Telegram update channel init
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)
	//OpenAI client init
	client := openai.NewClient(gAIToken)

	//Beginning of message processing
	go func() {
		for update := range updates {
			//Has been recieved callback
			if update.CallbackQuery != nil {
				if update.CallbackQuery.Data == "WHITELIST" || update.CallbackQuery.Data == "BLACKLIST" || update.CallbackQuery.Data == "INPROCESS" {
					keys, err = gRedisClient.Keys("ChatState:*").Result()
					if err != nil {
						SendToUser(gOwner, E12+err.Error(), ERROR)
						log.Fatalln(err)
					}
					//keys processing
					msgString = ""
					for _, key := range keys {
						jsonStr, err = gRedisClient.Get(key).Result()
						if err == redis.Nil {
							SendToUser(gOwner, E16+err.Error(), ERROR)
							log.Println(err)
							continue
						} else if err != nil {
							SendToUser(gOwner, E13+err.Error(), ERROR)
							log.Fatalln(err)
						} else {
							err = json.Unmarshal([]byte(jsonStr), &chatItem)
							if err != nil {
								SendToUser(gOwner, E14+err.Error(), ERROR)
								log.Fatalln(err)
							}
							if chatItem.AllowState == ALLOW && update.CallbackQuery.Data == "WHITELIST" {
								if chatItem.Type != "private" {
									msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
								} else {
									msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
								}
							}
							if chatItem.AllowState == DISALLOW && update.CallbackQuery.Data == "BLACKLIST" {
								if chatItem.Type != "private" {
									msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
								} else {
									msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
								}
							}
							if chatItem.AllowState == IN_PROCESS && update.CallbackQuery.Data == "INPROCESS" {
								if chatItem.Type != "private" {
									msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.Title + "\n"
								} else {
									msgString = msgString + "ID: " + strconv.FormatInt(chatItem.ChatID, 10) + " ~ " + chatItem.UserName + "\n"
								}
							}
						}
					}
					SendToUser(gOwner, msgString, SELECTCHAT)
				}
				if update.CallbackQuery.Data == "RESETTODEFAULTS" {
					err := gRedisClient.FlushAll().Err()
					if err != nil {
						SendToUser(gOwner, E10+err.Error(), ERROR)
						log.Fatalln(err)
					} else {
						SendToUser(gOwner, IM4, INFO)
						os.Exit(0)
					}
				}
				if update.CallbackQuery.Data == "FLUSHCACHE" {
					keys, err = gRedisClient.Keys("QuestState:*").Result()
					if err != nil {
						SendToUser(gOwner, E12+err.Error(), ERROR)
						log.Fatalln(err)
					}
					if len(keys) > 0 {
						msgString = "Кеш очищен\n"
						for _, key := range keys {
							err = gRedisClient.Del(key).Err()
							if err != nil {
								SendToUser(gOwner, E10+err.Error(), ERROR)
								log.Fatalln(err)
							}
						}
						SendToUser(gOwner, msgString+"Было удалено "+strconv.Itoa(len(keys))+" устаревших записей.", INFO)
					} else {
						SendToUser(gOwner, "Очищать нечего.", INFO)
					}
				}
				if update.CallbackQuery.Data == "RESTART" {
					SendToUser(gOwner, IM5, INFO)
					os.Exit(0)
				}
				if strings.Contains(update.CallbackQuery.Data, "ID:") {
					chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
					chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
					if err != nil {
						SendToUser(gOwner, E15+err.Error(), ERROR)
						log.Fatalln(err)
					}
					SendToUser(update.CallbackQuery.From.ID, "Выберите действие c чатом "+chatIDstr, TUNECHAT, chatID)
				}
				if strings.Contains(update.CallbackQuery.Data, "CLEAR_CONTEXT:") {
					chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
					chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
					if err != nil {
						SendToUser(gOwner, E15+err.Error(), ERROR)
						log.Fatalln(err)
					}
					jsonStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result()
					if err == redis.Nil {
						SendToUser(gOwner, E16+err.Error(), ERROR)
						log.Println(err)
						break
					} else if err != nil {
						SendToUser(gOwner, E13+err.Error(), ERROR)
						log.Fatalln(err)
					} else {
						err = json.Unmarshal([]byte(msgString), &chatItem)
						if err != nil {
							SendToUser(gOwner, E14+err.Error(), ERROR)
							log.Fatalln(err)
						}
						ChatMessages = chatItem.History
						jsonData, err = json.Marshal(ChatMessages)
						if err != nil {
							SendToUser(gOwner, E11+err.Error(), ERROR)
							log.Fatalln(err)
						}
						err = gRedisClient.Set("Dialog:"+chatIDstr, string(jsonData), 0).Err()
						if err != nil {
							SendToUser(gOwner, E10+err.Error(), ERROR)
							log.Fatalln(err)
						}
						SendToUser(chatID, "Контекст очищен!", NOTHING)
					}
				}
				if strings.Contains(update.CallbackQuery.Data, "GAME_IT_ALIAS") {
					chatIDstr := strings.Split(update.CallbackQuery.Data, " ")[1]
					chatID, err := strconv.ParseInt(chatIDstr, 10, 64)
					if err != nil {
						SendToUser(gOwner, E15+err.Error(), ERROR)
						log.Fatalln(err)
					}
					jsonStr, err = gRedisClient.Get("ChatState:" + chatIDstr).Result() //Читаем инфо от чате в БД
					if err == redis.Nil {
						SendToUser(gOwner, E16+err.Error(), ERROR)
						log.Println(err)
						break
					} else if err != nil {
						SendToUser(gOwner, E13+err.Error(), ERROR)
						log.Fatalln(err)
					} else {
						err = json.Unmarshal([]byte(msgString), &chatItem)
						if err != nil {
							SendToUser(gOwner, E14+err.Error(), ERROR)
							log.Fatalln(err)
						}
						ChatMessages = chatItem.History
						ChatMessages = append(ChatMessages, gITAlias...)
						jsonData, err = json.Marshal(ChatMessages)
						if err != nil {
							SendToUser(gOwner, E11+err.Error(), ERROR)
							log.Fatalln(err)
						}
						err = gRedisClient.Set("Dialog:"+chatIDstr, string(jsonData), 0).Err() //Записываем диалог в БД
						if err != nil {
							SendToUser(gOwner, E10+err.Error(), ERROR)
							log.Fatalln(err)
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
					jsonStr, err = gRedisClient.Get("QuestState:" + ansItem.CallbackID.String()).Result() //читаем состояние запрса из БД
					if err == redis.Nil {
						SendToUser(gOwner, E16+err.Error(), ERROR)
						log.Println(err)
					} else if err != nil {
						SendToUser(gOwner, E13+err.Error(), ERROR)
						log.Fatalln(err)
					} else {
						err = json.Unmarshal([]byte(jsonStr), &questItem)
						if err != nil {
							SendToUser(gOwner, E14+err.Error(), ERROR)
							log.Fatalln(err)
						}
						if questItem.State == QUEST_IN_PROGRESS {
							jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(questItem.ChatID, 10)).Result()
							if err == redis.Nil {
								SendToUser(gOwner, E16+err.Error(), ERROR)
								log.Println(err)
							} else if err != nil {
								SendToUser(gOwner, E13+err.Error(), ERROR)
								log.Fatalln(err)
							} else {
								err = json.Unmarshal([]byte(jsonStr), &chatItem)
								if err != nil {
									SendToUser(gOwner, E14+err.Error(), ERROR)
									log.Fatalln(err)
								}
							}
							switch ansItem.State { //Изменяем флаг доступа
							case ALLOW:
								{
									chatItem.AllowState = ALLOW
									SendToUser(gOwner, IM6, INFO)
									SendToUser(chatItem.ChatID, IM7, INFO)
								}
							case DISALLOW:
								{
									chatItem.AllowState = DISALLOW
									SendToUser(gOwner, IM8, INFO)
									SendToUser(chatItem.ChatID, IM9, INFO)
								}
							case BLACKLISTED:
								{
									chatItem.AllowState = BLACKLISTED
									SendToUser(gOwner, IM10, INFO)
									SendToUser(chatItem.ChatID, IM11, INFO)
								}
							}
							jsonData, err = json.Marshal(chatItem)
							if err != nil {
								SendToUser(gOwner, E11+err.Error(), ERROR)
								log.Fatalln(err)
							}
							err = gRedisClient.Set("ChatState:"+strconv.FormatInt(questItem.ChatID, 10), string(jsonData), 0).Err()
							if err != nil {
								SendToUser(gOwner, E10+err.Error(), ERROR)
								log.Fatalln(err)
							}
						}
					}
				}
				continue
			}
			if update.Message != nil {
				if update.Message.IsCommand() { //Begin command processing
					command := update.Message.Command()
					switch command {
					case "menu":
						if update.Message.Chat.ID == gOwner {
							SendToUser(gOwner, IM12, MENU)
						} else {
							SendToUser(update.Message.Chat.ID, IM12, USERMENU)
						}
					}
					continue
				}
				if update.Message.Text != "" { //Begin message processing
					jsonStr, err = gRedisClient.Get("ChatState:" + strconv.FormatInt(update.Message.Chat.ID, 10)).Result()
					if err == redis.Nil {
						log.Println(err) //Если записи в БД нет - формирруем новую запись
						chatItem.ChatID = update.Message.Chat.ID
						chatItem.BotState = RUN
						chatItem.AllowState = IN_PROCESS
						chatItem.UserName = update.Message.From.UserName
						chatItem.Type = update.Message.Chat.Type
						chatItem.Title = update.Message.Chat.Title
						chatItem.Model = openai.GPT3Dot5Turbo1106
						chatItem.Temperature = 1.1
						chatItem.Inity = 5
						chatItem.History = gHsOwner
						jsonData, err = json.Marshal(chatItem)
						if err != nil {
							SendToUser(gOwner, E11+err.Error(), ERROR)
							log.Fatalln(err)
						}
						err = gRedisClient.Set("ChatState:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err()
						if err != nil {
							SendToUser(gOwner, E10+err.Error(), ERROR)
							log.Fatalln(err)
						}
						if update.Message.Chat.Type == "private" {
							SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, update.Message.Chat.ID)
						} else {
							SendToUser(gOwner, "В группововм чате "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыли диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться в этом чате?", ACCESS, update.Message.Chat.ID)
						}
					} else if err != nil {
						SendToUser(gOwner, E13+err.Error(), ERROR)
						log.Fatalln(err)
					} else {
						err = json.Unmarshal([]byte(jsonStr), &chatItem)
						if err != nil {
							SendToUser(gOwner, E14+err.Error(), ERROR)
							log.Fatalln(err)
						}
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
										log.Println(err)
										ChatMessages = append(ChatMessages, chatItem.History...)
										if update.Message.Chat.Type == "private" { //Если текущий чат приватный
											ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.Text})
										} else { //Если текущи чат групповой записываем первое сообщение чата дополняя его именем текущего собеседника
											ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.From.FirstName + ": " + update.Message.Text})
										}
										jsonData, err = json.Marshal(ChatMessages)
										if err != nil {
											SendToUser(gOwner, E11+err.Error(), ERROR)
										}
										err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err() //Записываем диалог в БД
										if err != nil {
											SendToUser(gOwner, E10+err.Error(), ERROR) //Здесь могут быть всякие ошибки записи в БД
											log.Fatalln(err)
										}
									} else if err != nil {
										SendToUser(gOwner, E13+err.Error(), ERROR)
										log.Fatalln(err)
									} else { //Если диалог уже существует
										err = json.Unmarshal([]byte(msgString), &ChatMessages)
										if err != nil {
											SendToUser(gOwner, E14+err.Error(), ERROR)
											log.Fatalln(err)
										}
										if update.Message.Chat.Type == "private" { //Если текущий чат приватный
											ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.Text})

										} else { //Если текущи чат групповой дописываем сообщение чата дополняя его именем текущего собеседника
											ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: update.Message.From.FirstName + ": " + update.Message.Text})
										}
										jsonData, err = json.Marshal(ChatMessages)
										if err != nil {
											SendToUser(gOwner, E11+err.Error(), ERROR)
											log.Fatalln(err)
										}
										err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err() //Записываем диалог в БД
										if err != nil {
											SendToUser(gOwner, E10+err.Error(), ERROR) //Здесь могут быть всякие ошибки записи в БД
											log.Fatalln(err)
										}
									}
									for { //Здесь мы делаем паузу, позволяющую не отправлять промпты чаще чем раз в 20 секунд
										currentTime := time.Now()
										elapsedTime := currentTime.Sub(gLastRequest)
										time.Sleep(time.Second)
										if elapsedTime >= 20*time.Second {
											break
										}
									}
									action := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
									toBotFlag := false
									for _, name := range gBotNames { //Определим - есть ли в контексте последнего сообщения имя бота
										if (strings.Contains(strings.ToUpper(update.Message.Text), name)) || (update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == gBot.Self.ID) { //Если имя бота встречается
											toBotFlag = true
											break
										}
									}
									if update.Message.Chat.Type == "private" || toBotFlag {
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
												SendToUser(gOwner, E17+err.Error(), ERROR)
												log.Println(err)
												time.Sleep(20 * time.Second)
											} else {
												//log.Printf("Чат ID: %d Токенов использовано: %d", update.Message.Chat.ID, resp.Usage.TotalTokens)
												msg.Text = resp.Choices[0].Message.Content //Записываем ответ в сообщение
												break
											}
										}
									}
									ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: msg.Text})
									jsonData, err = json.Marshal(ChatMessages)
									if err != nil {
										SendToUser(gOwner, E11+err.Error(), ERROR)
										log.Fatalln(err)
									}
									err = gRedisClient.Set("Dialog:"+strconv.FormatInt(update.Message.Chat.ID, 10), string(jsonData), 0).Err()
									if err != nil {
										SendToUser(gOwner, E10+err.Error(), ERROR)
										log.Fatalln(err)
									}
									gBot.Send(msg)
								}
							case DISALLOW:
								{
									if update.Message.Chat.Type == "private" {
										SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.From.UserName+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться с этим пользователем?", ACCESS, update.Message.Chat.ID)
									} else {
										SendToUser(gOwner, "Пользователь "+update.Message.From.FirstName+" "+update.Message.Chat.Title+" открыл диалог.\nCообщение пользователя \n```\n"+update.Message.Text+"\n```\nРазрешите мне общаться в этом чате?", ACCESS, update.Message.Chat.ID)

									}
								}
							case BLACKLISTED:
								{
									if update.Message.Chat.Type == "private" {
										log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
									} else {
										log.Println("Запрос заблокированного диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
									}
								}
							case IN_PROCESS:
								{
									if update.Message.Chat.Type == "private" {
										log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.From.UserName + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
									} else {
										log.Println("Запрос диалога от " + update.Message.From.FirstName + " " + update.Message.Chat.Title + " " + strconv.FormatInt(update.Message.Chat.ID, 10))
									}
								}
							}
						}
					}
				}
			}
		}
	}()
	for {
		time.Sleep(time.Minute)
		rd := gRand.Intn(1000) + 1
		keys, err = gRedisClient.Keys("ChatState:*").Result()
		if err != nil {
			SendToUser(gOwner, E12+err.Error(), ERROR)
			log.Fatalln(err)
		}
		//keys processing
		msgString = ""
		for _, key := range keys {
			jsonStr, err = gRedisClient.Get(key).Result()
			if err != nil {
				SendToUser(gOwner, E13+err.Error(), ERROR)
				log.Fatalln(err)
			}
			err = json.Unmarshal([]byte(jsonStr), &chatItem)
			if err != nil {
				SendToUser(gOwner, E14+err.Error(), ERROR)
				log.Fatalln(err)
			}
			if rd <= chatItem.Inity && chatItem.AllowState == ALLOW {
				act := tgbotapi.NewChatAction(chatItem.ChatID, tgbotapi.ChatTyping)
				gBot.Send(act)
				for {
					currentTime := time.Now()
					elapsedTime := currentTime.Sub(gLastRequest)
					time.Sleep(time.Second)
					if elapsedTime >= 20*time.Second {
						break
					}
				}
				gLastRequest = time.Now() //Прежде чем формировать запрос, запомним текущее время
				ChatMessages = gIntFacts
				ansText := ""
				resp, err := client.CreateChatCompletion( //Формируем запрос к мозгам
					context.Background(),
					openai.ChatCompletionRequest{
						Model:       chatItem.Model,
						Temperature: chatItem.Temperature,
						Messages:    ChatMessages,
					},
				)
				if err != nil {
					SendToUser(gOwner, E17+err.Error(), ERROR)
					log.Println(err)
				} else {
					//log.Printf("Чат ID: %d Токенов использовано: %d", chatItem.ChatID, resp.Usage.TotalTokens)
					ansText = resp.Choices[0].Message.Content
					SendToUser(chatItem.ChatID, ansText, NOTHING)
				}
				msgString, err = gRedisClient.Get("Dialog:" + strconv.FormatInt(chatItem.ChatID, 10)).Result() //Пытаемся прочесть из БД диалог
				if err == redis.Nil {                                                                          //Если диалога в БД нет, формируем новый и записываем в БД
					ChatMessages = append(ChatMessages, chatItem.History...)
					ChatMessages = append(ChatMessages, gIntFacts...)
					ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: ansText})
					jsonData, err = json.Marshal(ChatMessages)
					if err != nil {
						SendToUser(gOwner, E11+err.Error(), ERROR)
						log.Fatalln(err)
					}
					err = gRedisClient.Set("Dialog:"+strconv.FormatInt(chatItem.ChatID, 10), string(jsonData), 0).Err() //Записываем диалог в БД
					if err != nil {
						SendToUser(gOwner, E10+err.Error(), ERROR)
						log.Fatalln(err)
					}
				} else if err != nil {
					SendToUser(gOwner, E13+err.Error(), ERROR)
					log.Fatalln(err)
				} else {
					err = json.Unmarshal([]byte(msgString), &ChatMessages)
					if err != nil {
						SendToUser(gOwner, E14+err.Error(), ERROR)
						log.Fatalln(err)
					}
					ChatMessages = append(ChatMessages, gIntFacts...)
					ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: ansText})
					jsonData, err = json.Marshal(ChatMessages)
					if err != nil {
						SendToUser(gOwner, E11+err.Error(), ERROR)
						log.Fatalln(err)
					}
					err = gRedisClient.Set("Dialog:"+strconv.FormatInt(chatItem.ChatID, 10), string(jsonData), 0).Err()
					if err != nil {
						SendToUser(gOwner, E10+err.Error(), ERROR)
						log.Fatalln(err)
					}
				}
			}
		}
	}
}

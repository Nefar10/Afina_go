package main

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
)

func init() {
	//Temporary variables
	var err error
	var owner int
	var db int

	//Logging level setup
	gVerboseLevel = 1

	//Randomize init
	gRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	//Read localization setting from OS env
	SetCurOperation("Environment initialization")
	switch os.Getenv(AFINA_LOCALE_IN_OS) {
	case "Ru":
		gLocale = 1
	case "En":
		gLocale = 0
	default:
		gLocale = 0
	}

	//Read bot API key from OS env
	gToken = os.Getenv(TOKEN_NAME_IN_OS)
	if gToken == "" {
		Log(E1[gLocale]+TOKEN_NAME_IN_OS+IM29[gLocale]+gCurProcName, CRIT, nil)
	}

	//Read owner's chatID from OS env
	owner, err = strconv.Atoi(os.Getenv(OWNER_IN_OS))
	if err != nil {
		Log(E2[gLocale]+OWNER_IN_OS+IM29[gLocale]+gCurProcName, CRIT, err)
	} else {
		gOwner = int64(owner) //Storing owner's chat ID in variable
		gChangeSettings = gOwner
	}

	//Telegram bot init
	gBot, err = tgbotapi.NewBotAPI(gToken)
	if err != nil {
		Log(E6[gLocale]+IM29[gLocale]+gCurProcName, CRIT, err)
	} else {
		if gVerboseLevel > 1 {
			gBot.Debug = true
		}
		Log(IM30[gLocale]+gBot.Self.UserName, NOERR, nil)
	}

	//Current dir init
	gDir, err = os.Getwd()
	if err != nil {
		SendToUser(gOwner, E8[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//Read redis connector options from OS env
	//Redis IP
	gRedisIP = os.Getenv(REDIS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, E3[gLocale]+REDIS_IN_OS+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//Redis password
	gRedisPass = os.Getenv(REDIS_PASS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, E4[gLocale]+REDIS_PASS_IN_OS+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//DB ID
	db, err = strconv.Atoi(os.Getenv(REDISDB_IN_OS))
	if err != nil {
		SendToUser(gOwner, E5[gLocale]+REDISDB_IN_OS+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
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
		SendToUser(gOwner, E9[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//Read OpenAI API token from OS env
	gAIToken = os.Getenv(AI_IN_OS)
	if gAIToken == "" {
		SendToUser(gOwner, E7[gLocale]+AI_IN_OS+IM29[gLocale]+gCurProcName, ERROR, 0)
	}

	//Read bot names from OS env
	gBotNames = strings.Split(os.Getenv(BOTNAME_IN_OS), ",")
	if gBotNames[0] == "" {
		gBotNames = gDefBotNames
		SendToUser(gOwner, IM1[gLocale]+BOTNAME_IN_OS+IM29[gLocale]+gCurProcName, INFO, 0)
	}

	//Bot naming prompt
	gHsName = [][]openai.ChatCompletionMessage{
		{
			{Role: openai.ChatMessageRoleUser, Content: "Your name is " + gBotNames[0] + "."},
			{Role: openai.ChatMessageRoleAssistant, Content: "Accepted! I'm " + gBotNames[0] + "."},
		},
		{
			{Role: openai.ChatMessageRoleUser, Content: "Тебя зовут " + gBotNames[0] + ""},
			{Role: openai.ChatMessageRoleAssistant, Content: "Принято! Мое имя " + gBotNames[0] + "."},
		},
	}

	//Read bot gender from OS env adn character comletion with gender information
	switch os.Getenv(BOTGENDER_IN_OS) {
	case "Male":
		gBotGender = MALE
	case "Female":
		gBotGender = FEMALE

	case "Neutral":
		gBotGender = NEUTRAL
	default:
		gBotGender = FEMALE
	}

	//Default chat states init
	gChatsStates = append(gChatsStates, ChatState{
		ChatID:      1,
		Model:       BASEGPTMODEL,
		Inity:       0,
		Temperature: 0.1,
		AllowState:  DISALLOW,
		UserName:    "All",
		BotState:    SLEEP,
		Type:        "private",
		History:     gHsNulled[gLocale],
		InterFacts:  0,
		Bstyle:      0,
		SetState:    NO_ONE,
		CharType:    ISTJ})
	gChatsStates = append(gChatsStates, ChatState{
		ChatID:      gOwner,
		Model:       BASEGPTMODEL,
		Inity:       0,
		Temperature: 0.5,
		AllowState:  ALLOW,
		UserName:    "Owner",
		BotState:    RUN,
		Type:        "private",
		History:     gHsNulled[gLocale],
		InterFacts:  0,
		Bstyle:      0,
		SetState:    NO_ONE,
		CharType:    ESFJ})

	//Storing default chat states to DB
	SetCurOperation(IM31[gLocale])
	for _, item := range gChatsStates {
		SetChatStateDB(item)
	}

	//OpenAI client init
	config := openai.DefaultConfig(gAIToken)
	config.BaseURL = "https://api.proxyapi.ru/openai/v1"
	gclient = openai.NewClientWithConfig(config)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	gModels = nil
	models, err := gclient.ListModels(ctx)
	if err != nil {
		SendToUser(gOwner, E18[gLocale], INFO, 1)
	} else {
		for _, model := range models.Models {
			if strings.Contains(strings.ToLower(model.ID), "gpt-4o") {
				gModels = append(gModels, model.ID)
			}
		}
	}
	gclient_is_busy = false

	//Send init complete message to owner
	SendToUser(gOwner, IM3[gLocale]+" "+IM13[gLocale], INFO, 0)
}

func ProcessMessages(update tgbotapi.Update) {
	switch {
	case update.MyChatMember != nil:
		ProcessMember(update)
	case update.CallbackQuery != nil:
		ProcessCallbacks(update)
	case update.Message != nil:
		switch {
		case update.Message.IsCommand():
			ProcessCommand(update)
		default:
			ProcessMessage(update)
		}
	}
}

func main() {
	var updateQueue []tgbotapi.Update
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)
	go func() {
		for update := range updates {
			updateQueue = append(updateQueue, update)
		}
	}()
	go func() {
		for {
			if len(updateQueue) > 0 {
				update := updateQueue[0]
				updateQueue = updateQueue[1:]
				gUpdatesQty = len(updateQueue)
				ProcessMessages(update)
			} else {
				time.Sleep(3000 * time.Millisecond)
			}
		}
	}()
	for {
		time.Sleep(time.Minute)
		ProcessInitiative()
	}
}

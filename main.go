package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
)

func init() {
	//Temporary variables
	var err error
	var owner int64
	var db int

	//Logging level setup
	gVerboseLevel = 1

	//Randomize init
	gRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	//Read localization setting from OS env
	SetCurOperation("Environment initialization", 0)
	switch os.Getenv(BOT_LOCALE_IN_OS) {
	case "Ru":
		gLocale = 1
	case "En":
		gLocale = 0
	default:
		gLocale = 0
	}
	SetCurOperation("Environment initialization | Prompts", 0)
	//Prompts init
	//saveCustomPrompts("prompts\\gHsReaction", gHsReaction)
	gHsGender, _ = loadCustomPrompts("prompts\\gHsGender.json")
	gConversationStyle, _ = loadCustomPrompts("prompts\\gConversationStyle.json")
	gIntFacts, _ = loadCustomPrompts("prompts\\gIntFacts.json")
	gHsGame, _ = loadCustomPrompts("prompts\\gHsGame.json")
	gHsReaction, _ = loadCustomPrompts("prompts\\gHsReaction.json")
	//log.Println(gHsReaction)
	//saveMsgs("msgs\\gBotReaction", gBotReaction)
	gErr, _ = loadMsgs("msgs\\gErr.json")
	gIm, _ = loadMsgs("msgs\\gIm.json")
	gMenu, _ = loadMsgs("msgs\\gMenu.json")
	gBotReaction, _ = loadMsgs("msgs\\gBotReaction.json")
	SetCurOperation("Environment initialization | T_API", 0)
	//Read bot API key from OS env
	gToken = os.Getenv(BOT_API_KEY_IN_OS)
	if gToken == "" {
		Log(gErr[1][gLocale]+BOT_API_KEY_IN_OS+gIm[29][gLocale]+GetCurOperation(), CRIT, nil)
	}
	SetCurOperation("Environment initialization | Owner", 0)
	//Read owner's chatID from OS env
	owner, err = strconv.ParseInt(os.Getenv(OWNER_IN_OS), 10, 64)
	if err != nil {
		Log(gErr[2][gLocale]+OWNER_IN_OS+gIm[29][gLocale]+GetCurOperation(), CRIT, err)
	} else {
		gOwner = int64(owner) //Storing owner's chat ID in variable
		gChangeSettings = 0
	}
	SetCurOperation("Environment initialization | Telegram connect", 0)
	//Telegram bot init
	gBot, err = tgbotapi.NewBotAPI(gToken)
	if err != nil {
		Log(gErr[6][gLocale]+gIm[29][gLocale]+GetCurOperation(), CRIT, err)
	} else {
		if gVerboseLevel > 1 {
			gBot.Debug = true
		}
		Log(gIm[30][gLocale]+gBot.Self.UserName, NOERR, nil)
	}

	//Current dir init
	//gDir, err = os.Getwd()
	//if err != nil {
	//	SendToUser(gOwner, gErr[8][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0)
	//}

	//Read redis connector options from OS env
	//Redis IP
	SetCurOperation("Environment initialization | Check DB", 0)
	gRedisIP = os.Getenv(REDIS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, gErr[3][gLocale]+REDIS_IN_OS+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0)
	}

	//Redis password
	gRedisPass = os.Getenv(REDIS_PASS_IN_OS)
	if gRedisPass == "" {
		SendToUser(gOwner, gErr[4][gLocale]+REDIS_PASS_IN_OS+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0)
	}

	//DB ID
	db, err = strconv.Atoi(os.Getenv(REDIS_DB_IN_OS))
	if err != nil {
		SendToUser(gOwner, gErr[5][gLocale]+REDIS_DB_IN_OS+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0)
	} else {
		gRedisDB = db //Storing DB ID
	}
	//gRedisDB = 9

	//Redis client init
	gRedisClient = redis.NewClient(&redis.Options{
		Addr:     gRedisIP,
		Password: gRedisPass,
		DB:       gRedisDB,
	})

	//Chek redis connection
	err = redisPing(*gRedisClient)
	if err != nil {
		SendToUser(gOwner, gErr[9][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0)
	}
	SetCurOperation("Environment initialization | Bot name", 0)
	//Read bot names from OS env
	gBotNames = strings.Split(os.Getenv(BOT_NAME_IN_OS), ",")
	if gBotNames[0] == "" {
		gBotNames = gDefBotNames
		SendToUser(gOwner, gIm[1][gLocale]+BOT_NAME_IN_OS+gIm[29][gLocale]+GetCurOperation(), MSG_INFO, 0)
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
	SetCurOperation("Environment initialization | Bot gender", 0)
	//Read bot gender from OS env adn character comletion with gender information
	switch os.Getenv(BOT_GENDER_IN_OS) {
	case "Male":
		gBotGender = MALE
	case "Female":
		gBotGender = FEMALE
	case "Neutral":
		gBotGender = NEUTRAL
	default:
		gBotGender = FEMALE
	}
	SetCurOperation("Environment initialization | AI settings", 0)
	//Read OpenAI API token from OS env and creating connections
	names := strings.Split(os.Getenv(AI_NAMES_IN_OS), ",")
	tokens := strings.Split(os.Getenv(AI_API_KEYS_IN_OS), ",")
	urls := strings.Split(os.Getenv(AI_URLS_IN_OS), ",")
	basemodels := strings.Split(os.Getenv(AI_BM_IN_OS), ",")
	gModels = nil
	for i, _ := range names {
		gAI = append(gAI, AI_params{AI_Name: names[i], AI_Token: tokens[i], AI_URL: urls[i], AI_BaseModel: basemodels[i]})
		config := openai.DefaultConfig(tokens[i])
		config.BaseURL = urls[i]
		gClient = append(gClient, openai.NewClientWithConfig(config))
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		models, err := gClient[i].ListModels(ctx)
		if err != nil {
			SendToUser(gOwner, gErr[18][gLocale], MSG_INFO, 1)
		} else {
			for _, model := range models.Models {
				if (strings.Contains(strings.ToLower(model.ID), "o1")) || (strings.Contains(strings.ToLower(model.ID), "4o")) ||
					(strings.Contains(strings.ToLower(model.ID), "o3")) || (strings.Contains(strings.ToLower(model.ID), "deep")) {
					gModels = append(gModels, AI_Models{AI_ID: i, AI_model_name: model.ID})
				}
			}
		}
	}
	if len(names) == 0 {
		SendToUser(gOwner, gErr[7][gLocale]+AI_API_KEYS_IN_OS+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0)
	}

	gClient_is_busy = false
	//Send init complete message to owner
	SendToUser(gOwner, gIm[3][gLocale]+" "+gIm[13][gLocale], MSG_INFO, 5)

}

func ProcessMessages(update tgbotapi.Update) {
	log.Println(update)
	switch {
	case update.MyChatMember != nil:
		ProcessMember(update)
	case update.CallbackQuery != nil:
		ProcessCallbacks(update)
	case update.Message != nil:
		switch {
		case update.Message.IsCommand():
			ProcessCommand(update)
		case update.Message.Location != nil:
			//ProcessLocation(update)
		default:
			ProcessMessage(update)
		}
	case update.EditedMessage != nil:
		switch {
		case update.EditedMessage.Location != nil:
			//ProcessLocation(update)
		default:
			{
			}
		}
	}
}

func main() {
	var updateQueue []tgbotapi.Update
	var updateMutex sync.Mutex
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)

	go func() {
		for update := range updates {
			updateMutex.Lock()
			updateQueue = append(updateQueue, update)
			updateMutex.Unlock()
		}
	}()
	go func() {
		for {
			if len(updateQueue) > 0 {
				updateMutex.Lock()
				update := updateQueue[0]
				updateQueue = updateQueue[1:]
				gUpdatesQty = len(updateQueue)
				updateMutex.Unlock()
				ProcessMessages(update)
			} else {
				time.Sleep(2000 * time.Millisecond)
			}
		}
	}()

	for {
		time.Sleep(time.Minute)
		ProcessInitiative()
		//ProcessNews()
	}
}

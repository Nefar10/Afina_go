package main

import (
	"context"
	"fmt"
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
	for _, arg := range os.Args[1:] {
		switch {
		case arg == "-vvv":
			gVerboseLevel = 3
		case arg == "-vv":
			gVerboseLevel = 2
		case arg == "-v":
			gVerboseLevel = 1
		default:
			gVerboseLevel = 0
		}

	}

	//Randomize init
	gRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	SetCurOperation("Environment initialization", 0)

	//Read localization setting from OS env
	SetCurOperation("Environment initialization | Determining locale data", 1)
	switch os.Getenv(BOT_LOCALE_IN_OS) {
	case "Ru":
		gLocale = 1
	case "En":
		gLocale = 0
	default:
		gLocale = 0
	}

	//Prompts init
	SetCurOperation("Environment initialization | Loading prompts", 1)
	//saveCustomPrompts("prompts\\gHsBasePrompt.json", gHsBasePrompt)
	gHsGender, _ = loadCustomPrompts("prompts/gHsGender.json")
	gHsBasePrompt, _ = loadCustomPrompts("prompts/gHsBasePrompt.json")
	gConversationStyle, _ = loadCustomPrompts("prompts/gConversationStyle.json")
	gIntFacts, _ = loadCustomPrompts("prompts/gIntFacts.json")
	gHsGame, _ = loadCustomPrompts("prompts/gHsGame.json")
	gHsReaction, _ = loadCustomPrompts("prompts/gHsReaction.json")
	gHsBasePrompt[0].Prompt[0][0].Content = fmt.Sprintf(gHsBasePrompt[0].Prompt[0][0].Content, VER)
	gHsBasePrompt[0].Prompt[1][0].Content = fmt.Sprintf(gHsBasePrompt[0].Prompt[1][0].Content, VER)
	//log.Println(gHsReaction)
	//saveMsgs("msgs\\gBotReaction", gBotReaction)
	gErr, _ = loadMsgs("msgs/gErr.json")
	gIm, _ = loadMsgs("msgs/gIm.json")
	gMenu, _ = loadMsgs("msgs/gMenu.json")
	gBotReaction, _ = loadMsgs("msgs/gBotReaction.json")
	SetCurOperation("Environment initialization | T_API", 1)

	//Read bot API key from OS env
	SetCurOperation("Environment initialization | Reading bot API key", 1)
	gToken = os.Getenv(BOT_API_KEY_IN_OS)
	if gToken == "" {
		Log(gErr[1][gLocale]+BOT_API_KEY_IN_OS+gIm[29][gLocale]+GetCurOperation(), CRIT, nil)
	}

	//Telegram bot init
	SetCurOperation("Environment initialization | Connecting to telegramm API", 1)
	gBot, err = tgbotapi.NewBotAPI(gToken)
	if err != nil {
		Log(gErr[6][gLocale]+gIm[29][gLocale]+GetCurOperation(), CRIT, err)
	} else {
		if gVerboseLevel > 1 {
			gBot.Debug = true
		}
		Log(gIm[30][gLocale]+gBot.Self.UserName, NOERR, nil)
	}

	//Read owner's chatID from OS env
	SetCurOperation("Environment initialization | Owner determining", 1)
	owner, err = strconv.ParseInt(os.Getenv(OWNER_IN_OS), 10, 64)
	if err != nil {
		Log(gErr[2][gLocale]+OWNER_IN_OS+gIm[29][gLocale]+GetCurOperation(), CRIT, err)
	} else {
		gOwner = int64(owner) //Storing owner's chat ID in variable
		gChangeSettings = 0
	}

	//Current dir init
	//gDir, err = os.Getwd()
	//if err != nil {
	//	SendToUser(gOwner, gErr[8][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0)
	//}

	//Read redis connector options from OS env
	//Redis IP
	SetCurOperation("Environment initialization | Reading DB connection credentials data", 1)
	gRedisIP = os.Getenv(REDIS_IN_OS)
	if gRedisIP == "" {
		SendToUser(gOwner, 0, gErr[3][gLocale]+REDIS_IN_OS+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	}

	//Redis password
	gRedisPass = os.Getenv(REDIS_PASS_IN_OS)
	if gRedisPass == "" {
		SendToUser(gOwner, 0, gErr[4][gLocale]+REDIS_PASS_IN_OS+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	}

	//DB ID
	db, err = strconv.Atoi(os.Getenv(REDIS_DB_IN_OS))
	if err != nil {
		SendToUser(gOwner, 0, gErr[5][gLocale]+REDIS_DB_IN_OS+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	} else {
		gRedisDB = db //Storing DB ID
	}
	//gRedisDB = 9

	//Redis client init
	SetCurOperation("Environment initialization | Connecting to DB", 1)
	gRedisClient = redis.NewClient(&redis.Options{
		Addr:     gRedisIP,
		Password: gRedisPass,
		DB:       gRedisDB,
	})

	//Chek redis connection
	err = redisPing(*gRedisClient)
	if err != nil {
		SendToUser(gOwner, 0, gErr[9][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	}

	SetCurOperation("Environment initialization | Determining bot's name", 1)
	//Read bot names from OS env
	gBotNames = strings.Split(os.Getenv(BOT_NAME_IN_OS), ",")
	if gBotNames[0] == "" {
		gBotNames = gDefBotNames
		SendToUser(gOwner, 0, gIm[1][gLocale]+BOT_NAME_IN_OS+gIm[29][gLocale]+GetCurOperation(), MSG_INFO, 0, false)
	}

	//Read bot gender from OS env adn character comletion with gender information
	SetCurOperation("Environment initialization | Determining bot's gender", 1)
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

	//Read OpenAI API token from OS env and creating connections
	SetCurOperation("Environment initialization | Loading AI data", 1)
	ainames := strings.Split(os.Getenv(AI_NAMES_IN_OS), ",")
	tokens := strings.Split(os.Getenv(AI_API_KEYS_IN_OS), ",")
	urls := strings.Split(os.Getenv(AI_URLS_IN_OS), ",")
	basemodels := strings.Split(os.Getenv(AI_BM_IN_OS), ",")
	gModels = nil
	for i, name := range ainames {
		gAI = append(gAI, AI_params{AI_Name: name, AI_Token: tokens[i], AI_URL: urls[i], AI_BaseModel: basemodels[i]})
		config := openai.DefaultConfig(tokens[i])
		config.BaseURL = urls[i]
		gClient = append(gClient, openai.NewClientWithConfig(config))
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		models, err := gClient[i].ListModels(ctx)
		if err != nil {
			SendToUser(gOwner, 0, gErr[18][gLocale], MSG_INFO, 1, false)
		} else {
			for _, model := range models.Models {
				if (strings.Contains(strings.ToLower(model.ID), "o1")) || (strings.Contains(strings.ToLower(model.ID), "4o")) ||
					(strings.Contains(strings.ToLower(model.ID), "o3")) || (strings.Contains(strings.ToLower(model.ID), "deep")) {
					gModels = append(gModels, AI_Models{AI_ID: i, AI_model_name: model.ID})
				}
			}
		}
	}
	if len(ainames) == 0 {
		SendToUser(gOwner, 0, gErr[7][gLocale]+AI_API_KEYS_IN_OS+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	}
	//gYaClient = yandexgpt.NewYandexGPTClientWithIAMToken("")

	gClient_is_busy = false
	//Send init complete message to owner
	SendToUser(gOwner, 0, fmt.Sprintf(gIm[3][gLocale]+" "+gIm[13][gLocale], VER), MSG_INFO, 5, true)

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
		case update.Message.Document != nil:
			ProcessDocument(update)
		case update.Message.Photo != nil:
			ProcessPhoto(update)
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
		//ProcessNews()
		time.Sleep(time.Minute)
		ProcessInitiative()
	}
}

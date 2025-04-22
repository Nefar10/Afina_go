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
	"github.com/sashabaranov/go-openai"
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
	switch os.Getenv(BotLocaleInOs) {
	case "Ru":
		gLocale = LocaleRu
	case "En":
		gLocale = LocaleEn
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
	gHsBasePrompt[0].Prompt[0][0].Content = fmt.Sprintf(gHsBasePrompt[0].Prompt[0][0].Content, Ver)
	gHsBasePrompt[0].Prompt[1][0].Content = fmt.Sprintf(gHsBasePrompt[0].Prompt[1][0].Content, Ver)
	//log.Println(gHsReaction)
	//saveMsgs("msgs\\gBotReaction", gBotReaction)
	gErr, _ = loadMsgs("msgs/gErr.json")
	gIm, _ = loadMsgs("msgs/gIm.json")
	gMenu, _ = loadMsgs("msgs/gMenu.json")
	gBotReaction, _ = loadMsgs("msgs/gBotReaction.json")
	SetCurOperation("Environment initialization | T_API", 1)

	//Read bot API key from OS env
	SetCurOperation("Environment initialization | Reading bot API key", 1)
	gToken = os.Getenv(BotApiKeyInOs)
	if gToken == "" {
		Log(gErr[1][gLocale]+BotApiKeyInOs+gIm[29][gLocale]+GetCurOperation(), ErrCritical, nil)
	}

	//Telegram bot init
	SetCurOperation("Environment initialization | Connecting to telegram API", 1)
	gBot, err = tgbotapi.NewBotAPI(gToken)
	if err != nil {
		Log(gErr[6][gLocale]+gIm[29][gLocale]+GetCurOperation(), ErrCritical, err)
	} else {
		if gVerboseLevel > 1 {
			gBot.Debug = true
		}
		Log(gIm[30][gLocale]+gBot.Self.UserName, ErrNo, nil)
	}

	//Read owner's chatID from OS env
	SetCurOperation("Environment initialization | Owner determining", 1)
	owner, err = strconv.ParseInt(os.Getenv(OwnerInOs), 10, 64)
	if err != nil {
		Log(gErr[2][gLocale]+OwnerInOs+gIm[29][gLocale]+GetCurOperation(), ErrCritical, err)
	} else {
		gOwner = owner //Storing owner's chat ID in variable
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
	gRedisIP = os.Getenv(RedisInOs)
	if gRedisIP == "" {
		SendToUser(gOwner, 0, gErr[3][gLocale]+RedisInOs+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}

	//Redis password
	gRedisPass = os.Getenv(RedisPassInOs)
	if gRedisPass == "" {
		SendToUser(gOwner, 0, gErr[4][gLocale]+RedisPassInOs+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}

	//DB ID
	db, err = strconv.Atoi(os.Getenv(RedisDbInOs))
	if err != nil {
		SendToUser(gOwner, 0, gErr[5][gLocale]+RedisDbInOs+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
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
		SendToUser(gOwner, 0, gErr[9][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}

	SetCurOperation("Environment initialization | Determining bot's name", 1)
	//Read bot names from OS env
	gBotNames = strings.Split(os.Getenv(BotNameInOs), ",")
	if gBotNames[0] == "" {
		gBotNames = gDefBotNames
		SendToUser(gOwner, 0, gIm[1][gLocale]+BotNameInOs+gIm[29][gLocale]+GetCurOperation(), MsgInfo, 0, false)
	}

	//Read bot gender from OS env adn character comletion with gender information
	SetCurOperation("Environment initialization | Determining bot's gender", 1)
	switch os.Getenv(BotGenderInOs) {
	case "Male":
		gBotGender = Male
	case "Female":
		gBotGender = Female
	case "Neutral":
		gBotGender = Neutral
	default:
		gBotGender = Female
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
	ainames := strings.Split(os.Getenv(AiNamesInOs), ",")
	tokens := strings.Split(os.Getenv(AiApiKeysInOs), ",")
	urls := strings.Split(os.Getenv(AiUrlsInOs), ",")
	basemodels := strings.Split(os.Getenv(AiBmInOs), ",")
	gModels = nil
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for i, name := range ainames {
		gAI = append(gAI, AiParams{AiName: name, AiToken: tokens[i], AiUrl: urls[i], AiBaseModel: basemodels[i]})
		config := openai.DefaultConfig(tokens[i])
		config.BaseURL = urls[i]
		gClient = append(gClient, openai.NewClientWithConfig(config))
		models, err := gClient[i].ListModels(ctx)
		if err != nil {
			SendToUser(gOwner, 0, gErr[ErrorCodeListModelsFailed][gLocale], MsgInfo, 1, false)
			continue
		} else {
			for _, model := range models.Models {
				modelIDLower := strings.ToLower(model.ID)
				if strings.Contains(modelIDLower, "o1") || strings.Contains(modelIDLower, "4o") ||
					strings.Contains(modelIDLower, "o3") || strings.Contains(modelIDLower, "deep") {
					gSysMutex.Lock()
					gModels = append(gModels, AiModels{AiId: i, AiModelName: model.ID})
					gSysMutex.Unlock()
				}
			}
		}
	}
	if len(ainames) == 0 {
		SendToUser(gOwner, 0, gErr[7][gLocale]+AiApiKeysInOs+gIm[29][gLocale]+GetCurOperation(), MsgError, 0, false)
	}
	//gYaClient = yandexgpt.NewYandexGPTClientWithIAMToken("")

	gClientIsBusy = false
	//Send init complete message to owner
	SendToUser(gOwner, 0, fmt.Sprintf(gIm[3][gLocale]+" "+gIm[13][gLocale], Ver), MsgInfo, 5, true)

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
	updateConfig.Timeout = UpdateConfigTimeout
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

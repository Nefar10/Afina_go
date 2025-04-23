package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

const (
	// BotLocaleInOs Localization data en/ru etc. in OS environment.
	BotLocaleInOs = "AFINA_LOCALE"
	// BotApiKeyInOs Telegram Bot API key in OS environment.
	BotApiKeyInOs = "TB_API_KEY"
	// AiApiKeysInOs List of API keys in OS environment.
	AiApiKeysInOs = "AI_KEYS"
	// AiNamesInOs List of AI names in OS environment.
	AiNamesInOs = "AI_NAMES"
	// AiUrlsInOs List of AI API URLs in OS environment.
	AiUrlsInOs = "AI_URLS"
	// AiBmInOs List of AI base models in OS environment.
	AiBmInOs = "AI_BM"
	// OwnerInOs Owner's telegram chat ID in OS environment.
	OwnerInOs = "OWNER"
	// BotNameInOs List of bot names in OS environment.
	BotNameInOs = "AFINA_NAMES"
	// BotGenderInOs Bot's gender data Male/Female/Neutral in OS environment.
	BotGenderInOs = "AFINA_GENDER"

	// RedisInOs Redis ip address and port data in OS environment.
	RedisInOs = "REDIS_IP"
	// RedisDbInOs Number of redis DB in OS environment.
	RedisDbInOs = "REDIS_DB" //Number DB in redis
	// RedisPassInOs Password for redis in OS environment.
	RedisPassInOs = "REDIS_PASS" //Pass for redis

	// UpdateConfigTimeout Some telegram client settings.
	UpdateConfigTimeout = 60

	// Male Bot's gender.
	Male = 1
	// Female Bot's gender.
	Female = 2
	// Neutral Bot's gender.
	Neutral = 0

	// ChatDisallow Chat status is disallowed. Denied access to communicate with bot.
	ChatDisallow = 0
	// ChatInProcess Chat status is undefined. Request to communicate in progress. No access to communicate with bot.
	ChatInProcess = 1
	// ChatAllow Chat status is allowed. Allow access to communicate with bot.
	ChatAllow = 2
	// ChatBlacklist Chat status is blocked. All access to bot is blocked.
	ChatBlacklist = 3

	// BotSleep Bot's status is sleep - bot sleeps and only views all messages.
	BotSleep = 0
	// BotRun Bot's status is run - bot lists and interacts with chat.
	BotRun = 1

	// QuestInProgress Status of user's request from bot - is not solved yet.
	QuestInProgress = 1
	// QuestSolved Status of user's request from bot - is solved and may be cleared with clear cache procedure.
	QuestSolved = 2 //Quest is solved

	// MsgNothing Do nothing flag for send to user procedure.
	MsgNothing = 0
	// MenuGetAccess Shows to owner menu of chat rights settings.
	MenuGetAccess = 1
	// MenuShowMenu Shows admin menu in chat.
	MenuShowMenu = 2
	// MenuShowUserMenu Shows user menu in chat.
	MenuShowUserMenu = 3
	// MenuSelChat Shows menu with list of chats for selecting.
	MenuSelChat = 4
	// MenuTuneChat Shows settings menu for selected chat.
	MenuTuneChat = 5
	// MsgError Sends error message to chat.
	MsgError = 6
	// MsgInfo Sends info message to chat.
	MsgInfo = 7
	// MenuTuneUser Shows settings menu for current chat.
	MenuTuneUser = 8
	// MenuSetIf Shows interest facts menu for selecting existing themes.
	MenuSetIf = 9
	// MenuSetStyle Shows menu list of styles for selecting bot's conversation style in current chat.
	MenuSetStyle = 10
	// MenuSetModel Shows menu list of AI models for selecting model to usage in current chat.
	MenuSetModel = 11
	// MenuShowChar Shows list of bot's character for selecting bot's conversation character in current chat.
	MenuShowChar = 12
	// MenuSetChar Runs set char function.
	MenuSetChar = 13
	// MenuSetTimezone Show list of time zones for selecting one in current chat.
	MenuSetTimezone = 14 //Select time zone

	// MenuAccessNo No access to menu.
	MenuAccessNo = 1
	// MenuAccessDefault Default user menu.
	MenuAccessDefault = 2
	// MenuAccessModerator Moderator menu.
	MenuAccessModerator = 4
	// MenuAccessAdmin Administrator menu.
	MenuAccessAdmin = 8
	// MenuAccessOwner Owner menu .
	MenuAccessOwner = 16

	// LocaleRu Russian locale.
	LocaleRu = 1
	// LocaleEn English locale.
	LocaleEn = 0

	// TaskNew Task state is new.
	TaskNew = 0
	// TaskInProgress Task state is in progress.
	TaskInProgress = 1
	// TaskIsDone Task state is done.
	TaskIsDone = 2
	// TaskError Task state in error
	TaskError = -1

	// ParamNoOne Nothing to do.
	ParamNoOne = 0
	// ParamHistory Set up history for chat.
	ParamHistory = 1
	// ParamTemperature Set up bot's temperature for chat
	ParamTemperature = 2
	// ParamInitiative Set up bot's initiative for chat
	ParamInitiative = 4
	// ParamBotCharacter Set up bot's character
	ParamBotCharacter = 5

	// ErrNo No have errors
	ErrNo = 0
	// Err Have an error
	Err = 1
	// ErrCritical Have a critical error
	ErrCritical = 2

	// DoNothing Call bot's function base answer.
	DoNothing = 0
	// DoCalculate Call bot's function calculate.
	DoCalculate = 1
	// DoShowMenu Call bot's function show menu.
	DoShowMenu = 2
	// DoShowHistory Call bot's function export chat history.
	DoShowHistory = 3
	// DoClearHistory Call bot's function clear context.
	DoClearHistory = 4
	// DoGame Call bot's function run game.
	DoGame = 5
	// DoReadSite Call bot's function read site.
	DoReadSite = 6
	// DoSearch Call bot's function web search.
	DoSearch = 7

	// Ver Code version.
	Ver = "0.33.212"

	// ErrorCodeListModelsFailed Fail to get list of ai-models.
	ErrorCodeListModelsFailed = 18

	// CHARACTER TYPES
	ISTJ = 1  // (Инспектор): Ответственный, организованный, практичный.
	ISFJ = 2  // (Защитник): Заботливый, внимательный, преданный.
	INFJ = 3  // (Советчик): Интуитивный, идеалистичный, глубоко чувствующий.
	INTJ = 4  // (Стратег): Аналитичный, независимый, целеустремленный.
	ISTP = 5  // (Мастер): Практичный, гибкий, решающий проблемы.
	ISFP = 6  // (Артист): Творческий, чувствительный, ценящий красоту.
	INFP = 7  // (Мечтатель): Идеалистичный, эмоциональный, ищущий смысл.
	INTP = 8  // (Мыслитель): Логичный, независимый, теоретический.
	ESTP = 9  // (Динамик): Энергичный, практичный, любящий приключения.
	ESFP = 10 // (Исполнитель): Общительный, веселый, любящий жизнь.
	ENFP = 11 // (Вдохновитель): Творческий, энтузиаст, ищущий новые возможности.
	ENTP = 12 // (Новатор): Инноватор, любящий обсуждения и новые идеи.
	ESTJ = 13 // (Руководитель): Организованный, практичный, ориентированный на результат.
	ESFJ = 14 // (Помощник): Заботливый, общительный, стремящийся помочь другим.
	ENFJ = 15 // (Наставник): Вдохновляющий, заботливый, умеющий вести за собой.
	ENTJ = 16 // (Командир): Решительный, стратегический, лидер по натуре.
)

//***TYPES***

// CharTypes Char types structure
type CharTypes struct {
	ID          byte
	Type        string
	Description []string
}

// ChatState Chat's structure for storing chat options in DB and operate with chat item
type ChatState struct {
	ChatID      int64                          `json:"chatid"`      //Chat ID
	UserName    string                         `json:"username"`    //Username
	Title       string                         `json:"title"`       //Group chat title
	AllowState  int                            `json:"allowstate"`  //Communicate allowing flag
	BotState    int                            `json:"botstate"`    //Bot state flag
	MenuLevel   int                            `json:"menulevel"`   //Set of access rights
	Type        string                         `json:"type"`        //Chat type - private/group/supergroup/channel
	Model       string                         `json:"model"`       //GPT model selected
	Temperature float32                        `json:"temperature"` //Bot's creativity
	History     []openai.ChatCompletionMessage `json:"history"`     //Current chat prompts
	InterFacts  int                            `json:"interfacts"`  //Interesting facts prompt
	Inity       int                            `json:"inty"`        //Bot's initiative
	CStyle      int                            `json:"cstyle"`      //Conversation style
	SetState    byte                           `json:"setstate"`    //While change setting
	CharType    byte                           `json:"chartype"`    //Character type by myers-Briggs
	Profession  byte                           `json:"profession"`  //Profession simulate
	TimeZone    int                            `json:"timezone"`    //Current time zone
	AiId        int                            `json:"aiid"`        //AI service ID
	Geo         UserLocation                   `json:"geo"`         //Chat location
	ContextLen  int                            `json:"contextlen"`  //Context length for chat
}

// QuestState Quest operating structure for processing access rights
type QuestState struct {
	ChatID     int64     `json:"chatid"`     //Chat ID
	CallbackID uuid.UUID `json:"callbackid"` //Quest uuid
	Question   int       `json:"question"`   //Question type
	State      int       `json:"state"`      //Solving state
	Time       time.Time `json:"time"`       //Current time
}

// Answer Callback answers structure for solving quests
type Answer struct {
	CallbackID uuid.UUID `json:"callbackid"` //Callback ID
	State      int       `json:"state"`      //Answer's state
}

// UserLocation Location structure for storing and processing user location
type UserLocation struct {
	Longitude            float32 `json:"longitude"`
	Latitude             float32 `json:"latitude"`
	HorizontalAccuracy   int     `json:"horizontalaccuracy"`
	LivePeriod           int     `json:"liveperiod"`
	Heading              int     `json:"heading"`
	ProximityAlertRadius int     `json:"proximityalertradius"`
}

// CustomPrompt Structure for custom prompts
type CustomPrompt struct {
	Id       int                              `json:"id"`
	Category string                           `json:"category"`
	Name     string                           `json:"name"`
	Prompt   [][]openai.ChatCompletionMessage `json:"prompt"`
}

// AiParams Structure for storing AI service data
type AiParams struct {
	AiName      string `json:"ainame"`
	AiToken     string `json:"aitoken"`
	AiUrl       string `json:"aiurl"`
	AiBaseModel string `json:"aibasemodel"`
}

// AiModels Structure for storing model name
type AiModels struct {
	AiId        int    `json:"aiid"`
	AiModelName string `json:"aimodelname"`
}

// ParsedData Structure for storing web content
type ParsedData struct {
	Content []string `json:"content"`
}

// Task Structure for storing task data
type Task struct {
	ID                int64         `json:"id"`
	Description       string        `json:"description"`
	Prompt            string        `json:"prompt"`
	CreatedAt         time.Time     `json:"createdat"`
	LastExecutedAt    time.Time     `json:"lastexecutedat"`
	ChatID            int64         `json:"chatid"`
	IsRecurring       bool          `json:"isrecurring"`
	Interval          time.Duration `json:"interval"`
	NextExecutionTime time.Time     `json:"nextexecutiontime"`
	Priority          int           `json:"priority"`
	State             int           `json:"state"`
}

//***VARIABLES***

// gCharTypes Collection of char types
var gCharTypes = [16]CharTypes{
	{
		ID:   ISTJ,
		Type: "ISTJ",
		Description: []string{"(Inspector): Responsible, organized, practical.",
			"(Инспектор): Ответственный, организованный, практичный."},
	},
	{
		ID:   ISFJ,
		Type: "ISFJ",
		Description: []string{"(Protector): Caring, attentive, devoted.",
			"(Защитник): Заботливый, внимательный, преданный."},
	},
	{
		ID:   INFJ,
		Type: "INFJ",
		Description: []string{"(Advisor): Intuitive, idealistic, deeply feeling",
			"(Советчик): Интуитивный, идеалистичный, глубоко чувствующий."},
	},
	{
		ID:   INTJ,
		Type: "INTJ",
		Description: []string{"(Strategist): Analytical, independent, goal-oriented.",
			"(Стратег): Аналитичный, независимый, целеустремленный."},
	},
	{
		ID:   ISTP,
		Type: "ISTP",
		Description: []string{"(Master): Practical, flexible, problem-solving.",
			"(Мастер): Практичный, гибкий, решающий проблемы."},
	},
	{
		ID:   ISFP,
		Type: "ISFP",
		Description: []string{"(Artist): Creative, sensitive, appreciating beauty.",
			"(Артист): Творческий, чувствительный, ценящий красоту."},
	},
	{
		ID:   INFP,
		Type: "INFP",
		Description: []string{"(Dreamer): Idealistic, emotional, searching for meaning.",
			"(Мечтатель): Идеалистичный, эмоциональный, ищущий смысл."},
	},
	{
		ID:   INTP,
		Type: "INTP",
		Description: []string{"(Thinker): Logical, independent, theoretical.",
			"(Мыслитель): Логичный, независимый, теоретический."},
	},
	{
		ID:   ESTP,
		Type: "ESTP",
		Description: []string{"(Dynamo): Energetic, practical, adventure-loving.",
			"(Динамик): Энергичный, практичный, любящий приключения."},
	},
	{
		ID:   ESFP,
		Type: "ESFP",
		Description: []string{"(Executor): Sociable, cheerful, life-loving.",
			"(Исполнитель): Общительный, веселый, любящий жизнь."},
	},
	{
		ID:   ENFP,
		Type: "ENFP",
		Description: []string{"(Inspirer): Creative, enthusiastic, seeking new opportunities.",
			"(Вдохновитель): Творческий, энтузиаст, ищущий новые возможности."},
	},
	{
		ID:   ENTP,
		Type: "ENTP",
		Description: []string{"(Innovator): Innovative, loving discussions and new ideas.",
			"(Новатор): Инноватор, любящий обсуждения и новые идеи."},
	},
	{
		ID:   ESTJ,
		Type: "ESTJ",
		Description: []string{"(Leader): Organized, practical, results-oriented.",
			"(Руководитель): Организованный, практичный, ориентированный на результат."},
	},
	{
		ID:   ESFJ,
		Type: "ESFJ",
		Description: []string{"(Helper): Caring, sociable, striving to help others.",
			"(Помощник): Заботливый, общительный, стремящийся помочь другим."},
	},
	{
		ID:   ENFJ,
		Type: "ENFJ",
		Description: []string{"(Mentor): Inspiring, caring, able to lead others.",
			"(Наставник): Вдохновляющий, заботливый, умеющий вести за собой."},
	},
	{
		ID:   ENTJ,
		Type: "ENTJ",
		Description: []string{"(Commander): Decisive, strategic, a natural leader.",
			"(Командир): Решительный, стратегический, лидер по натуре."},
	},
}

// gTimeZones Collection of time zones
var gTimeZones = []string{
	"- UTC-12:00",       //0
	"- UTC-11:00",       //1
	"- UTC-10:00",       //2
	"- UTC-09:00",       //3
	"- UTC-08:00",       //4
	"- UTC-07:00",       //5
	"- UTC-06:00",       //6
	"- UTC-05:00",       //7
	"- UTC-04:00",       //8
	"- UTC-03:00",       //9
	"- UTC-02:00",       //10
	"- UTC-01:00",       //11
	"- UTC±00:00 (UTC)", //12
	"- UTC+01:00",       //13
	"- UTC+02:00",       //14
	"- UTC+03:00",       //15
	"- UTC+04:00",       //16
	"- UTC+05:00",       //17
	"- UTC+06:00",       //18
	"- UTC+07:00",       //19
	"- UTC+08:00",       //20
	"- UTC+09:00",       //21
	"- UTC+10:00",       //22
	"- UTC+11:00",       //23
	"- UTC+12:00",       //24
	"- UTC+13:00",       //25
	"- UTC+14:00",       //26
}

// gErr Error messages
var gErr [][2]string

// gIm Info messages
var gIm [][2]string

// gMenu Menus
var gMenu [][2]string

// gBotReaction Bot reactions
var gBotReaction [][2]string

// gHsGender Prompts collection for determining gender
var gHsGender []CustomPrompt

// gConversationStyle Prompts collection for determining conversation style
var gConversationStyle []CustomPrompt

// gIntFacts Prompts collection for self initiative actions
var gIntFacts []CustomPrompt

// gHsGame Prompts collection for gaming
var gHsGame []CustomPrompt

// gHsReaction Prompts collection with bot's reaction in public chats
var gHsReaction []CustomPrompt

// gHsBasePrompt Base prompts for all conversation styles
var gHsBasePrompt []CustomPrompt

// gHsName Prompts collection for bot naming
var gHsName [][]openai.ChatCompletionMessage

// gBot Pointer to initialized bot
var gBot *tgbotapi.BotAPI

// gClient AI services clients
var gClient []*openai.Client

// gYaClient
// var gYaClient *yandexgpt.YandexGPTClient

// gClientIsBusy Flag request to API is active
var gClientIsBusy bool

// gLastRequest Time of last request to AI service
var gLastRequest time.Time

// gLocale Current localization
var gLocale byte

// gToken Bot API tokens
var gToken string

// gOwner Bot's owner chat ID for send confirmations
var gOwner int64

// gBotNames Bot names for calling he in group chats
var gBotNames []string

// gBotGender Bot's gender
var gBotGender int

// gRedisIP DB server address and port
var gRedisIP string

// gRedisDB DB number in redis
var gRedisDB int

// gRedisPass Password for redis DB
var gRedisPass string

// gRedisClient Pointer for redis client
var gRedisClient *redis.Client

// gAI Collection of AI services parameters
var gAI []AiParams

// gRand New Random numbers generator
var gRand *rand.Rand

// gCurProcName Current process name for logging and debugging
var gCurProcName string

// gUpdatesQty Telegram bot updates buffer length
var gUpdatesQty int

// gModels Collection of reached AI models
var gModels []AiModels

// gVerboseLevel Logging level
var gVerboseLevel byte

// gBotLocation Server or owner's home location
var gBotLocation UserLocation

// gChangeSettings Chat ID where settings changes now
var gChangeSettings int64

// gAIMutex Mutex for AI request synchronization
var gAIMutex sync.Mutex

// gSysMutex Mutex for local variables access synchronization
var gSysMutex sync.Mutex

// gScheduler Collection of tasks
var gScheduler []Task

// gDefBotNames Bot's default name
var gDefBotNames = []string{"Athena", "Афина"}

// gDefChatState Global variable for storing default chat settings
var gDefChatState = ChatState{
	ChatID:      0,
	BotState:    BotRun,
	AllowState:  ChatInProcess,
	UserName:    "NoName",
	Type:        "NoType",
	Title:       "NoTitle",
	AiId:        0,
	Model:       "gpt-4o-mini",
	Temperature: 0.5,
	Inity:       0,
	History:     nil,
	InterFacts:  0,
	CStyle:      0,
	SetState:    ParamNoOne,
	CharType:    ESTJ,
	TimeZone:    15,
	Geo: UserLocation{
		Longitude:            0,
		Latitude:             0,
		HorizontalAccuracy:   0,
		LivePeriod:           0,
		Heading:              0,
		ProximityAlertRadius: 0},
	ContextLen: 64000,
}

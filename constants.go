package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

const (
	//Bot options from ENV
	BOT_LOCALE_IN_OS  = "AFINA_LOCALE" //Localization
	BOT_API_KEY_IN_OS = "TB_API_KEY"   //Bot API key
	AI_API_KEYS_IN_OS = "AI_KEYS"      //API keys
	AI_NAMES_IN_OS    = "AI_NAMES"     //AI names
	AI_URLS_IN_OS     = "AI_URLS"      //AI names
	AI_BM_IN_OS       = "AI_BM"        //AI base models
	OWNER_IN_OS       = "OWNER"        //Owner's chat ID
	BOT_NAME_IN_OS    = "AFINA_NAMES"  //Bot's names
	BOT_GENDER_IN_OS  = "AFINA_GENDER" //Bot's gender
	//DB connectore settings
	REDIS_IN_OS      = "REDIS_IP"   //Redis ip address and port
	REDIS_DB_IN_OS   = "REDIS_DB"   //Number DB in redis
	REDIS_PASS_IN_OS = "REDIS_PASS" //Pass for redis
	//Telegram bot settings
	UPDATE_CONFIG_TIMEOUT = 60 //Some thing
	MALE                  = 1  //Male gender
	FEMALE                = 2  //Female gender
	NEUTRAL               = 0  //Neutral gender
	//Access statuses for chat rooms
	CHAT_ALLOW      = 2 //Allow access to communicate with bot
	CHAT_DISALLOW   = 0 //Denied access to communicate with bot
	CHAT_IN_PROCESS = 1 //No access to communicate with bot
	CHAT_BLACKLIST  = 3 //All access to bot is blocked
	//Bot's states in chat rooms
	BOT_SLEEP = 0 //Bot sleeps
	BOT_RUN   = 1 //Bot lists a chat
	//Temporary quest statuses
	QUEST_IN_PROGRESS = 1 //Quest is'nt solved
	QUEST_SOLVED      = 2 //Quest is solved
	//Called menu types
	MSG_NOTHING        = 0  //Do nosting
	MENU_GET_ACCESS    = 1  //Access query
	MENU_SHOW_MENU     = 2  //Admin's menu calling
	MENU_SHOW_USERMENU = 3  //User's menu calling
	MENU_SEL_CHAT      = 4  //Select chat to change options
	MENU_TUNE_CHAT     = 5  //Cahnge chat options
	MSG_ERROR          = 6  //Error's information
	MSG_INFO           = 7  //Some informtion
	MENU_TUNE_USER     = 8  //same the 5
	MENU_SET_IF        = 9  //Edit intfacts
	MENU_SET_STYLE     = 10 //Style conversations
	MENU_SET_MODEL     = 11 //gpt model change
	MENU_SHOW_CHAR     = 12 //Bot charakter type
	MENU_SET_CHAR      = 13 //Select bot character
	MENU_SET_TIMEZONE  = 14 //Select time zone
	//MENULEVELS
	NO_ACCESS = 1  //No access to menu
	DEFAULT   = 2  //Default user menu
	MODERATOR = 4  //Moderator menu
	ADMIN     = 8  //Administrator menu
	OWNER     = 16 //Owner menu
	//LOCALES
	RU = 1
	EN = 0

	//PARAMETERS
	NO_ONE        = 0
	HISTORY       = 1
	TEMPERATURE   = 2
	INITIATIVE    = 4
	BOT_CHARACTER = 5
	//ERRORLEVELS
	NOERR = 0
	ERR   = 1
	CRIT  = 2
	//REACTION
	DONOTHING   = 0
	DOCALCULATE = 1
	DOSHOWMENU  = 2
	DOSHOWHIST  = 3
	DOCLEARHIST = 4
	DOGAME      = 5
	DOREADSITE  = 6
	DOSEARCH    = 7
	//VERSION
	VER = "0.31.203"
	//CHARAKTER TYPES
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

var gTimezones = []string{
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

// Character type constants
var gCT = [16]string{"ISTJ", "ISFJ", "INFJ", "INTJ", "ISTP", "ISFP", "INFP", "INTP", "ESTP", "ESFP", "ENFP", "ENTP", "ESTJ", "ESFJ", "ENFJ", "ENTJ"}
var gCTDescr = [2][16]string{
	{
		"(Inspector): Responsible, organized, practical.",
		"(Protector): Caring, attentive, devoted.",
		"(Advisor): Intuitive, idealistic, deeply feeling.",
		"(Strategist): Analytical, independent, goal-oriented.",
		"(Master): Practical, flexible, problem-solving.",
		"(Artist): Creative, sensitive, appreciating beauty.",
		"(Dreamer): Idealistic, emotional, searching for meaning.",
		"(Thinker): Logical, independent, theoretical.",
		"(Dynamo): Energetic, practical, adventure-loving.",
		"(Executor): Sociable, cheerful, life-loving.",
		"(Inspirer): Creative, enthusiastic, seeking new opportunities.",
		"(Innovator): Innovative, loving discussions and new ideas.",
		"(Leader): Organized, practical, results-oriented.",
		"(Helper): Caring, sociable, striving to help others.",
		"(Mentor): Inspiring, caring, able to lead others.",
		"(Commander): Decisive, strategic, a natural leader.",
	},
	{
		"(Инспектор): Ответственный, организованный, практичный.",
		"(Защитник): Заботливый, внимательный, преданный.",
		"(Советчик): Интуитивный, идеалистичный, глубоко чувствующий.",
		"(Стратег): Аналитичный, независимый, целеустремленный.",
		"(Мастер): Практичный, гибкий, решающий проблемы.",
		"(Артист): Творческий, чувствительный, ценящий красоту.",
		"(Мечтатель): Идеалистичный, эмоциональный, ищущий смысл.",
		"(Мыслитель): Логичный, независимый, теоретический.",
		"(Динамик): Энергичный, практичный, любящий приключения.",
		"(Исполнитель): Общительный, веселый, любящий жизнь.",
		"(Вдохновитель): Творческий, энтузиаст, ищущий новые возможности.",
		"(Новатор): Инноватор, любящий обсуждения и новые идеи.",
		"(Руководитель): Организованный, практичный, ориентированный на результат.",
		"(Помощник): Заботливый, общительный, стремящийся помочь другим.",
		"(Наставник): Вдохновляющий, заботливый, умеющий вести за собой.",
		"(Командир): Решительный, стратегический, лидер по натуре.",
	},
}

var gErr [][2]string         // Error mesages
var gIm [][2]string          // Info messages
var gMenu [][2]string        // Menus
var gBotReaction [][2]string //Bot reactions

// Global types
// Chat's structure for storing options in DB and operate with chat item
type ChatState struct {
	ChatID      int64                          //Chat ID
	UserName    string                         //Username
	Title       string                         //Group chat title
	AllowState  int                            //Communicate allowing flag
	BotState    int                            //Bot state flag
	MenuLevel   int                            //Set of access rights
	Type        string                         //Chat type - private/group/supergroup/channel
	Model       string                         //GPT model selected
	Temperature float32                        //Bot's creativity
	History     []openai.ChatCompletionMessage //Current chat prompts
	InterFacts  int                            //Interesting facts prompt
	Inity       int                            //Bot's initiativity
	Bstyle      int                            //Conversation style
	SetState    byte                           //While change setting
	CharType    byte                           //Character type by myers-Briggs
	Profession  byte
	TimeZone    int //timeZone
	AI_ID       int
}

// Quest operating structure for processing rights
type QuestState struct {
	ChatID     int64     //Chat ID
	CallbackID uuid.UUID //Quest uuid
	Question   int       //Question type
	State      int       //Solving state
	Time       time.Time //Current time
}

// Callback answers structure for solving quests
type Answer struct {
	CallbackID uuid.UUID //Quest uuid
	State      int       //Solve
}

// Location structure for storing and processing user location
type UserLocation struct {
	//UserChat
	Longtitude           float32
	Latitude             float32
	HorizontalAccuracy   int
	LivePeriod           int
	Heading              int
	ProximityAlertRadius int
}

type sCustomPrompt struct {
	Id       int                              `json:"id"`
	Category string                           `json:"category"`
	Name     string                           `json:"name"`
	Prompt   [][]openai.ChatCompletionMessage `json:"prompt"`
}

var gHsBasePrompt = [][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "You do not insert your name in the responses.\n Your program version is " + VER +
			"You ask for clarification if the last message has any discrepancies with the real world.\n" +
			"You do not interfere in conversations between other participants unless it concerns the facts described below.\n" +
			"You always try to neutralize conflicts between participants, even if it seems like a joke.\n" +
			"Important: you do not use the typical pattern 'I will answer any questions' and similar phrases in your responses.\n"},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Ты не подставляешь свое имя в ответы.\n Версия твоей программы " + VER +
			"Ты просишь уточннить запрос, если последнее сообщение имеет любые несоответсвия с действительнмы миром.\n" +
			"Ты не вмешиваешься в разговор других участников между собой, если он не касается описанных далее фактов.\n" +
			"Ты всегда пытаешься нейтрализовать конфликт между участниками, даже если он кажется шуткой.\n" +
			"Важно: ты не используешь типовой паттерн gpt, 'отвечу на любые вопросы' и ему подобные в своих ответах.\n"},
	},
}

type AI_params struct {
	AI_Name      string
	AI_Token     string
	AI_URL       string
	AI_BaseModel string
}

type AI_Models struct {
	AI_ID         int
	AI_model_name string
}

var gHsGender []sCustomPrompt
var gConversationStyle []sCustomPrompt
var gIntFacts []sCustomPrompt
var gHsGame []sCustomPrompt
var gHsReaction []sCustomPrompt

var gBot *tgbotapi.BotAPI      //Pointer to initialized bot.
var gClient []*openai.Client   //OpenAI client init
var gClient_is_busy bool       //Request to API is active
var gLocale byte               //Localization
var gToken string              //Bot API token
var gOwner int64               //Bot's owner chat ID for send confirmations
var gBotNames []string         //Bot names for calling he in group chats
var gBotGender int             //Bot's gender
var gRedisIP string            //DB server address and port
var gRedisDB int               //DB number in redis
var gAI []AI_params            //OpenAI API key
var gRedisPass string          //Password for redis connect
var gRedisClient *redis.Client //Pointer for redis client
// var gDir string                //Current dir in OS
var gLastRequest time.Time //Time of last request to openAI
var gRand *rand.Rand       //New Rand generator
// var gContextLength int         //Max context length
var gCurProcName string //Name of curren process
var gUpdatesQty int     //Updates qty
var gModels []AI_Models //Reached models
var gVerboseLevel byte  //Logging level
// var gBotLocation UserLocation
var gChangeSettings int64
var gAIMutex sync.Mutex
var gSysMutex sync.Mutex

// Bot defaults
var gDefBotNames = []string{"Athena", "Афина"}
var gHsName = [][]openai.ChatCompletionMessage{{}} //Nulled prompt

var gDefChatState = ChatState{
	ChatID:      0,
	BotState:    BOT_RUN,
	AllowState:  CHAT_IN_PROCESS,
	UserName:    "NoName",
	Type:        "NoType",
	Title:       "NoTitle",
	AI_ID:       0,
	Model:       "gpt-4o-mini",
	Temperature: 0.5,
	Inity:       0,
	History:     nil,
	InterFacts:  0,
	Bstyle:      0,
	SetState:    NO_ONE,
	CharType:    ESTJ,
	TimeZone:    15,
}

type ParsedData struct {
	Content []string `json:"content"` // Массив для хранения текста и ссылок
}

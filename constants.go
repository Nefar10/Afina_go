package main

import (
	"math/rand"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

const (
	//Bot options from ENV
	TOKEN_NAME_IN_OS = "TB_API_KEY"    //Bot API key
	AI_IN_OS         = "AFINA_API_KEY" //OpenAI API key
	OWNER_IN_OS      = "OWNER"         //Owner's chat ID
	BOTNAME_IN_OS    = "AFINA_NAMES"   //Bot's names
	BOTGENDER_IN_OS  = "AFINA_GENDER"  //Bot's gender
	//DB connectore settings
	REDIS_IN_OS      = "REDIS_IP"   //Redis ip address and port
	REDISDB_IN_OS    = "REDIS_DB"   //Number DB in redis
	REDIS_PASS_IN_OS = "REDIS_PASS" //Pass for redis
	//Telegram bot settings
	UPDATE_CONFIG_TIMEOUT = 60 //Some thing
	MALE                  = 1  //Male gender
	FEMALE                = 2  //Female gender
	NEUTRAL               = 0  //Neutral gender
	//Access statuses for chat rooms
	ALLOW       = 2 //Allow access to communicate with bot
	DISALLOW    = 0 //Denied access to communicate with bot
	IN_PROCESS  = 1 //No access to communicate with bot
	BLACKLISTED = 3 //All access to bot is blocked
	//Bot's states in chat rooms
	SLEEP = 0 //Bot sleeps
	RUN   = 1 //Bot lists a chat
	//Temporary quest statuses
	QUEST_IN_PROGRESS = 1 //Quest is'nt solved
	QUEST_SOLVED      = 2 //Quest is solved
	//Called menu types
	NOTHING      = 0 //Do nosting
	ACCESS       = 1 //Access query
	MENU         = 2 //Admin's menu calling
	USERMENU     = 3 //User's menu calling
	SELECTCHAT   = 4 //Select chat to change options
	TUNECHAT     = 5 //Cahnge chat options
	ERROR        = 6 //Error's information
	INFO         = 7 //Some informtion
	TUNECHATUSER = 8 //same the 5
	INTFACTS     = 9 //Edit intfacts
	//MENULEVELS
	NO_ACCESS = 1  //No access to menu
	DEFAULT   = 2  //Default user menu
	MODERATOR = 4  //Moderator menu
	ADMIN     = 8  //Administrator menu
	OWNER     = 16 //Owner menu
	//ERRORS
	E1  = " Telegram bot API key not forund in OS environment "
	E2  = " Owner's chat ID not found or not valid in os environment "
	E3  = " DB server IP address and port not found in OS environment "
	E4  = " DB password not found in OS environment "
	E5  = " DB ID not forind or not valid in OS environment "
	E6  = " Telegram bot initialization error "
	E7  = " OpenAI API tocken not found in OS environment "
	E8  = " Error initializing work dir "
	E9  = " DB connection error "
	E10 = " Error writting to DB "
	E11 = " Error json marshaling "
	E12 = " Error reading keys from DB "
	E13 = " Error reading key value from DB "
	E14 = " Error json unmarshaling "
	E15 = " Error convetring string to int "
	E16 = " Unknown Error "
	E17 = " ChatCompletion error: %v\n "
	//INFO MESSAGES
	IM0  = " Program has been stoped "
	IM1  = " Bot name(s) not found or not valid in OS environment.\n Name Afina will be used. "
	IM2  = " Bot gender not found or not valid in OS environment.\n Neutral gender will be used. "
	IM3  = " I'm back! "
	IM4  = " All DB data has been remowed. I'll reboot now "
	IM5  = " I'll be back "
	IM6  = " Access granted "
	IM7  = " I was allowed to communicate with you! "
	IM8  = " Access denied "
	IM9  = " I apologize, but to continue the conversation, it is necessary to subscribe. "
	IM10 = " Access bocked "
	IM11 = " Congratulations! You have been added to the pranksters list! "
	IM12 = " Please select what needs to be done. "
	IM13 = "Current version is 0.3.0"
	IM14 = " Choose a topic. "
	IM15 = " Topic has been changed. "
)

// Global types
// Chat's structure for storing options in DB and operate with them
type ChatState struct {
	ChatID      int64                          //Chat ID
	UserName    string                         //Username
	Title       string                         //Group chat title
	AllowState  int                            //Communicate allowing flag
	BotState    int                            //Bot state flag
	MenuLevel   int                            //Set of access rights
	Type        string                         //Chat type - private/group/supergroup
	Model       string                         //GPT model selected
	Temperature float32                        //Bot's creativity
	History     []openai.ChatCompletionMessage //Current chat prompts
	IntFacts    []openai.ChatCompletionMessage //Interesting facts prompt
	Inity       int                            //Bot's initiativity
}

// Quest operating structure
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

// Presetted prompts
// Nulled prompt
var gHsNulled = []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: ""}}

// Default prompt
var gHsOwner = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль универсального персонального асисстента. Зови себя - Афина."},
	{Role: openai.ChatMessageRoleAssistant, Content: "Здравствуйте. Поняла, можете называть меня Афина. Я Ваш универсальный ассистент."}}

// Game IT-alias prompt
var gITAlias = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleUser, Content: "Давай поиграем в IT Элиас. Ты будешь в роли ведущего. Правила следующие:\n" +
		"1) Ты загадываешь сложный термин из области IT поддержки и рассказываешь - что это такое не используя однокоренных слов\n" +
		"2) Ты не должен называть загаданный термин, пока он не будет отгадан или не закончатся попытки.\n" +
		"3) У нас есть три попытки, чтобы отгадать очередной загаданный термин. После каждой нашей попытки ты сообщаешь о количестве оставшихся попыток.\n" +
		"4) После завершения каждого тура ты предлагаешь продолжить игру."},
	{Role: openai.ChatMessageRoleAssistant, Content: "Понял. Я буду загазывать различные термины из области IT поддержки и не буду называть их."}}
var gIntFactsGen = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleUser, Content: "Расскажи только один необычный и интересный факт.\n" +
		"Важно начать с фразы 'Интересный факт!'."}}
var gIntFactsSci = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleUser, Content: "Расскажи только один необычный и интересный факт из области естественных наук.\n" +
		"Важно начать с фразы 'Интересный факт!'."}}
var gIntFactsIT = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleUser, Content: "Расскажи только один необычный и интересный факт из области IT.\n" +
		"Важно начать с фразы 'Интересный факт!'."}}
var gIntFactsAuto = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleUser, Content: "Расскажи только один необычный и интересный факт про автомобилии или гонки или компьютерные игры. \n" +
		"Важно начать с фразы 'Интересный факт!' и максимально самокритично озвучивать рекорды."}}

var gBot *tgbotapi.BotAPI //Pointer to initialized bot.
// OpenAI client init
var gclient *openai.Client
var gclient_is_busy bool       //Request to API is active
var gToken string              //Bot API token
var gOwner int64               //Bot's owner chat ID for send confirmations
var gBotNames []string         //Bot names for calling he in group chats
var gBotGender int             //Bot's gender
var gChatsStates []ChatState   //Default chat states initialization
var gRedisIP string            //DB server address and port
var gRedisDB int               //DB number in redis
var gAIToken string            //OpenAI API key
var gRedisPass string          //Password for redis connect
var gRedisClient *redis.Client //Pointer for redis client
var gDir string                //Current dir in OS
var gLastRequest time.Time     //Time of last request to openAI
var gRand *rand.Rand           //New Rand generator
var gContextLength int         //Max context length

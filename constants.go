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
	AFINA_LOCALE_IN_OS = "AFINA_LOCALE"  //Localization
	TOKEN_NAME_IN_OS   = "TB_API_KEY"    //Bot API key
	AI_IN_OS           = "AFINA_API_KEY" //OpenAI API key
	OWNER_IN_OS        = "OWNER"         //Owner's chat ID
	BOTNAME_IN_OS      = "AFINA_NAMES"   //Bot's names
	BOTGENDER_IN_OS    = "AFINA_GENDER"  //Bot's gender
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
	//LOCALES
	RU = 1
	EN = 0
)

// ERRORS
var E1 = [2]string{" Telegram bot API key not forund in OS environment ", " Не найден API ключ телеграмм бота в переменных окружения "}
var E2 = [2]string{" Owner's chat ID not found or not valid in os environment ", " Не найден ID чата владельца в переменных окружения "}
var E3 = [2]string{" DB server IP address and port not found in OS environment ", " Адрес сервера базы данных не найден в переменных окружения "}
var E4 = [2]string{" DB password not found in OS environment ", " Пароль к базе данных не найден в переменных окружения "}
var E5 = [2]string{" DB ID not forind or not valid in OS environment ", " Идентификатор базы не найден в переменных окружения "}
var E6 = [2]string{" Telegram bot initialization error ", " Ошибка инициализации телеграмм бота "}
var E7 = [2]string{" OpenAI API tocken not found in OS environment ", " API ключ OpenAI не найден в перменных окружения "}
var E8 = [2]string{" Error initializing work dir ", " Ошибка инициализации рабочей директории "}
var E9 = [2]string{" DB connection error ", " Ошибка подключения к базе данных "}
var E10 = [2]string{" Error writting to DB ", " Ошибка записи в базу данных "}
var E11 = [2]string{" Error json marshaling ", " Ошибка преобразования в json "}
var E12 = [2]string{" Error reading keys from DB ", " Ошибка чтения ключа из базы данных "}
var E13 = [2]string{" Error reading key value from DB ", " Ошибка чтения знаяения ключа из базы данных "}
var E14 = [2]string{" Error json unmarshaling ", " Ошибка парсинга Json "}
var E15 = [2]string{" Error convetring string to int ", " Ошибка преобразования строки в число "}
var E16 = [2]string{" Unknown Error ", " Неизвестная ошибка "}
var E17 = [2]string{" ChatCompletion error: %v\n ", " Ошибка обработки запроса к нейросети: %v\n"}

// INFO MESSAGES
var IM0 = [2]string{" Process has been stoped ", " Процесс был остановлен "}
var IM1 = [2]string{" Bot name(s) not found or not valid in OS environment.\n Name Afina will be used. ", " Имя бота не найдено или не корректно в переменных окружения.\n Будет использовано имя Afina. "}
var IM2 = [2]string{" Bot gender not found or not valid in OS environment.\n Neutral gender will be used. ", " Пол бота не найден или некорректен среди переменных окружения.\n Будет использован средний род. "}
var IM3 = [2]string{" I'm back! ", " Я снова с Вами! "}
var IM4 = [2]string{" All DB data has been remowed. I'll reboot now ", " Все данные в бахе данных будут удалены. Проиводится перезагрузка "}
var IM5 = [2]string{" I'll be back ", " Я еще вернусь "}
var IM6 = [2]string{" Access granted ", " Доступ разрешен "}
var IM7 = [2]string{" I was allowed to communicate with you! ", " Мне было разрешено с вами общаться "}
var IM8 = [2]string{" Access denied ", " Доступ запрещен "}
var IM9 = [2]string{" I apologize, but to continue the conversation, it is necessary to subscribe. ", " Простите, но для продолжения общения необходимо оформить подписку. "}
var IM10 = [2]string{" Access bocked ", " Доступ заблокирован "}
var IM11 = [2]string{" Congratulations! You have been added to the pranksters list! ", " Поздравляю! Вы были добавлены в список проказников! "}
var IM12 = [2]string{" Please select what needs to be done. ", " Пожалуйста, выберите, что необходимо выполнить. "}
var IM13 = [2]string{" Current version is 0.3.2", " Текущая версия 0.3.2"}
var IM14 = [2]string{" Choose a topic. ", " Выберите тему "}
var IM15 = [2]string{" Topic has been changed. ", " Тема изменена "}
var IM16 = [2]string{" Write - as soon as you are ready to start the game. ", " Пишите - как только будете готовы начать игру. "}

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
var gHsOwner = [2][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "Hi! You are playing the role of a universal personal assistant. Call yourself - Athena."},
		{Role: openai.ChatMessageRoleAssistant, Content: "Hello! Got it, you can call me Athena. I am your universal assistant."},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль универсального персонального ассистента. Зови себя - Афина."},
		{Role: openai.ChatMessageRoleAssistant, Content: "Здравствуйте. Поняла, можете называть меня Афина. Я Ваш универсальный ассистент."},
	},
}

// Game IT-alias prompt
var gITAlias = [2][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "Let’s play IT Charades. You’ll take on the role of the host. The rules are as follows:\n" +
			"1) You’ll think of a complex term from the IT support realm and explain what it is without using any root words.\n" +
			"2) You mustn't reveal the term until it’s guessed or we run out of attempts.\n" +
			"3) We have three chances to guess the chosen term. After each of our guesses, you’ll let us know how many attempts we have left.\n" +
			"4) After each round, you’ll ask if we want to keep the ball rolling."},
		{Role: openai.ChatMessageRoleAssistant, Content: "Got it. I will think of various terms from the IT support field and I won’t reveal them."},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Давай поиграем в IT Элиас. Ты будешь в роли ведущего. Правила следующие:\n" +
			"1) Ты загадываешь сложный термин из области IT поддержки и рассказываешь - что это такое не используя однокоренных слов\n" +
			"2) Ты не должен называть загаданный термин, пока он не будет отгадан или не закончатся попытки.\n" +
			"3) У нас есть три попытки, чтобы отгадать очередной загаданный термин. После каждой нашей попытки ты сообщаешь о количестве оставшихся попыток.\n" +
			"4) После завершения каждого тура ты предлагаешь продолжить игру."},
		{Role: openai.ChatMessageRoleAssistant, Content: "Понял. Я буду загазывать различные термины из области IT поддержки и не буду называть их."},
	},
}
var gIntFactsGen = [2][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "Tell me just one unusual and interesting fact. It's important to start with the phrase 'Interesting fact!'."},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Расскажи только один необычный и интересный факт. Важно начать с фразы 'Интересный факт!'."},
	},
}
var gIntFactsSci = [2][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "Tell me just one unusual and interesting fact from the natural sciences. It's important to start with the phrase 'Interesting fact!'."},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Расскажи только один необычный и интересный факт из области естественных наук. Важно начать с фразы 'Интересный факт!'."},
	},
}
var gIntFactsIT = [2][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "Tell me just one unusual and interesting fact from the field of IT. It's important to start with the phrase 'Interesting fact!'."},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Расскажи только один необычный и интересный факт из области IT. Важно начать с фразы 'Интересный факт!'."},
	},
}
var gIntFactsAuto = [2][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "Tell me just one unusual and interesting fact about cars, racing, or video games. It's important to start with the phrase 'Interesting fact!' and to mention records in a self-deprecating manner."},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Расскажи только один необычный и интересный факт про автомобилии или гонки или компьютерные игры. Важно начать с фразы 'Интересный факт!' и максимально самокритично озвучивать рекорды."},
	},
}
var gBot *tgbotapi.BotAPI //Pointer to initialized bot.
// OpenAI client init
var gclient *openai.Client
var gclient_is_busy bool       //Request to API is active
var gLocale byte               //Localization
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
var gCurProcName string        //Name of curren process

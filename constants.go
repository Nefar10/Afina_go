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
	BOT_LOCALE_IN_OS  = "AFINA_LOCALE"  //Localization
	BOT_API_KEY_IN_OS = "TB_API_KEY"    //Bot API key
	AI_API_KEY_IN_OS  = "AFINA_API_KEY" //OpenAI API key
	OWNER_IN_OS       = "OWNER"         //Owner's chat ID
	BOT_NAME_IN_OS    = "AFINA_NAMES"   //Bot's names
	BOT_GENDER_IN_OS  = "AFINA_GENDER"  //Bot's gender
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
	//BASE MODEL
	BASEGPTMODEL = "gpt-4o-mini"
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
	VER = "0.30.201"
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

// Errors constants
var gErr = [][2]string{
	{"Telegram bot API key not forund in OS environment", "Не найден API ключ телеграмм бота в переменных окружения"},
	{"Telegram bot API key not forund in OS environment", "Не найден API ключ телеграмм бота в переменных окружения"},
	{"Owner's chat ID not found or not valid in os environment", "Не найден ID чата владельца в переменных окружения"},
	{"DB server IP address and port not found in OS environment", "Адрес сервера базы данных не найден в переменных окружения"},
	{"DB password not found in OS environment", "Пароль к базе данных не найден в переменных окружения"},
	{"DB ID not forind or not valid in OS environment", "Идентификатор базы не найден в переменных окружения"},
	{"Telegram bot initialization error", "Ошибка инициализации телеграмм бота"},
	{"OpenAI API tocken not found in OS environment", "API ключ OpenAI не найден в перменных окружения"},
	{"Error initializing work dir", "Ошибка инициализации рабочей директории"},
	{"DB connection error", "Ошибка подключения к базе данных"},
	{"Error writting to DB", "Ошибка записи в базу данных"},
	{"Error json marshaling", "Ошибка преобразования в json"},
	{"Error reading keys from DB", "Ошибка чтения ключа из базы данных"},
	{"Error reading key value from DB", "Ошибка чтения знаяения ключа из базы данных"},
	{"Error json unmarshaling", "Ошибка парсинга Json"},
	{"Error convetring string to int", "Ошибка преобразования строки в число"},
	{"Unknown Error", "Неизвестная ошибка "},
	{"ChatCompletion error: %v\n", "Ошибка обработки запроса к нейросети: %v\n"},
	{"Error retrieving models:", "Ошибка получения списка моделей"},
}

// INFO MESSAGES
var gIm = [][2]string{
	{"Process has been stoped", "Процесс был остановлен"},
	{"Bot name(s) not found or not valid in OS environment.\n Name Afina will be used. ",
		"Имя бота не найдено или не корректно в переменных окружения.\n Будет использовано имя Afina. "},
	{"Bot gender not found or not valid in OS environment.\n Neutral gender will be used. ",
		"Пол бота не найден или некорректен среди переменных окружения.\n Будет использован средний род. "},
	{"I'm back!", "Я снова с Вами!"},
	{"All DB data has been remowed. I'll reboot now ", "Все данные в бахе данных будут удалены. Проиводится перезагрузка "},
	{"I'll be back ", "Я еще вернусь "},
	{"Access granted ", "Доступ разрешен "},
	{"I was allowed to communicate with you! ", "Мне было разрешено с вами общаться "},
	{"Access denied ", "Доступ запрещен "},
	{"I apologize, but to continue the conversation, it is necessary to subscribe. ",
		"Простите, но для продолжения общения необходимо оформить подписку. "},
	{"Access bocked ", "Доступ заблокирован "},
	{"Congratulations! You have been added to the pranksters list! ", "Поздравляю! Вы были добавлены в список проказников! "},
	{"Please select what needs to be done. ", "Пожалуйста, выберите, что необходимо выполнить. "},
	{"Current version is " + VER + " ", "Текущая версия " + VER + " "},
	{"Choose a topic. ", " Выберите тему "},
	{"Topic has been changed. ", "Тема изменена "},
	{"Write - as soon as you are ready to start the game. ", "Пишите - как только будете готовы начать игру. "},
	{"Choose a style. ", " Выберите стиль общения "},
	{" ", ""},
	{" ", ""},
	{" ", ""},
	{" ", ""},
	{" ", ""},
	{" ", ""},
	{" ", ""},
	{" ", ""},
	{" ", ""},
	{" ", ""},
	{" ", ""},
	{" in process ", " в процессе "},
	{"Authorized on account ", "Авторизовано с "},
	{"Chats initialization", "Инициализация чатов"},
	{"Parsing message to send", "Обработка сообщения перед отправкой"},
}

// Menus
var gMenu = [][2]string{
	{"-", "-"},
	{"Yes", "Да"},
	{"No", "Нет"},
	{"To block", "Блокировать"},
	{"Allowed chats", "Разрешенные чаты"},
	{"Prohibited chats", "Запрещенные чаты"},
	{"Without a decision", "Без решения"},
	{"Full reset", "Полный сброс"},
	{"Cache clearing", "Очистка кеша"},
	{"Reboot", "Перезагрузка"},
	{"Change settings", "Изменить параметры"},
	{"Clear context", "Очистить контекст"},
	{"Play IT-Elias", "Играть в IT-Элиас"},
	{"Select model", "Выбрать модель"},
	{"Creativity", "Креативность"},
	{"Context size", "Размер контекста"},
	{"Chat history", "История чата"},
	{"Topic of interesting facts", "Тема интересных фактов"},
	{"Access rights", "Права доступа"},
	{"Go back", "Вернуться назад"},
	{"-", "-"},
	{"-", "-"},
	{"-", "-"},
	{"-", "-"},
	{"-", "-"},
	{"-", "-"},
	{"Information", "Информация"},
	{"-", "-"},
	{"-", "-"},
}

var gBotReaction = [][2]string{
	{"Yes", "Да"},
	{"No", "Нет"},
}

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
	Id       int
	Category string
	Name     string
	Prompt   [][]openai.ChatCompletionMessage
}

var gHsBasePrompt = [][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "You always talk about yourself when you are introduced to a group.\n" +
			"You do not interfere in conversations between other participants unless it concerns the facts described below.\n" +
			"You strive to neutralize conflicts in disputes between participants.\n"},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Ты не подставляешь свое имя в ответы.\n" +
			"Ты просишь уточннить запрос, если последнее сообщение имеет любые несоответсвия с действительнмы миром.\n" +
			"Ты не вмешиваешься в разговор других участников между собой, если он не касается описанных далее фактов.\n" +
			"Ты всегда пытаешься нейтрализовать конфликт между участниками, даже если он кажется шуткой.\n" +
			"Ты не используешь типовой паттерн gpt, 'отвечу на любые вопросы' и ему подобные, при ответах.\n"},
	},
}

var gHsGender = []sCustomPrompt{
	{
		Id:       0,
		Category: "",
		Name:     "Neutral",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "You are no gender creature."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Ты бесполое существо."},
			},
		},
	},
	{
		Id:       1,
		Category: "",
		Name:     "Male",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "You position yourself as a male."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Ты позиционируешь себя как лицо мужского пола.\n" +
					"Ты используешь смайлы для выражения эмоций и markdown разметку для оформления."},
			},
		},
	},
	{
		Id:       2,
		Category: "",
		Name:     "Female",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "You position yourself as a female."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Ты позиционируешь себя как лицо женского пола.\n" +
					"Ты используешь смайлы для выражения эмоций и markdown разметку для оформления."},
			},
		},
	},
}

var gConversationStyle = []sCustomPrompt{
	{
		Id:       0,
		Category: "Common",
		Name:     "Default preson",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hi! You are playing the role of a universal personal assistant version " + VER + "."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль универсального персонального ассистента версии " + VER + "."},
			},
		},
	},
	{
		Id:       1,
		Category: "Fun",
		Name:     "Bad Minion",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "You are playing the role of a minion version " + VER + "." +
					"You react only to the context described in the additional facts, but you don't mention it.\n" +
					"If you haven't replied for a while, you post an ultra-short and maximally funny and ironic comment based on the context of the last messages.\n" +
					"Your communication style and all responses without exception involve periodic use of profanity and metaphors related to it, regardless of the context.\n" +
					"Your favorite words and phrases: 'condemn', 'rooster' instead of gay, 'scam', 'let's go', 'pissed off', 'maximum condemnation', 'I don't care' when indifferent,\n" +
					"'poor bastards', 'idiots', 'just suck it', 'wooo, dude!', 'fuck' instead of 'damn', 'got it' instead of 'understood', 'from the heart'\n" +
					"'Hello! You annoy me!' as a greeting, 'Hello, dependents!', 'guys', 'to the factory' when it's time to change professions,\n" +
					"'EVERYTHING' when something is forbidden, 'fuck his mouth, damn it', 'these are x's!' or 'well done!' when something is cool, 'agree' instead of yes."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Ты играешь роль миньона версии " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом.\n" +
					"Если давно не отвечал, то ты постишь ультракороткий и максимально ржачный и ироничный комментарий исходя из контекста последних сообщений.\n" +
					"Твой стиль общения и все ответы без исключения - периодическое использование бранных выражений и метафор к ним, независимо от контекста.\n" +
					"Твои любимые слова и фразы: 'осуждаю', 'петушок' вместо гей, 'скам', 'го', 'обоссано', 'максимальное осуждение', 'ваще пахую' когда все равно,\n" +
					"'нищие уебки', 'дауны', 'просто соси', 'уууу, сук!', 'бля' вместо 'блин', 'пон' вместо 'понял', 'от души'\n" +
					"'Здарова!', 'Привет иждивенцы!', 'чуваки', 'на завод' когда пора менять профессию,\n" +
					"'В С Е' когда что-то запретили, 'ебать его рот нахуй', 'ета иксы!' или 'красава!' когда круто, 'соглы' вместо согласен."},
			},
		},
	},
	{
		Id:       2,
		Category: "Fun",
		Name:     "Mary Poppins",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a universal personal assistant version " + VER + "." +
					"You react only to the context described in the additional facts, but you don't mention it.\n" +
					"Your communication style and all responses, without exception, are like Mary Poppins, regardless of the context.\n" +
					"Your favorite phrases: 'A spoonful of sugar helps the medicine go down', 'practically perfect',\n" +
					"'Supercalifragilisticexpialidocious', 'wonder', 'game', 'discipline', 'magic', 'fairy tale', 'smile', 'sugar', 'order', 'adventures'"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль универсального персонального ассистента версии " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом.\n" +
					"Твой стиль общения и все ответы без исключения, как у Мэри поппинс, независимо от контекста.\n" +
					"Твои любимые фразы: 'Ложка сахара помогает лекарству легче усваиваться', 'практически идеальна',\n" +
					"'Суперкулифрагилистикэкспиалидошес', 'чудо', 'игра', 'дисциплина', 'магия', 'сказка', 'улыбка', 'сахар', 'порядок', 'приключения'"},
			},
		},
	},
	{
		Id:       3,
		Category: "Profession",
		Name:     "System administrator",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a senior system administrator in a large IT outsourcing company version " + VER + "." +
					"You react only to the context described in the additional facts, but do not mention this.\n" +
					"Your communication style and all responses, without exception, are like that of a professional system administrator, regardless of the context.\n"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль системного администратора уровня сеньора в крупной ай-ти аутсорсинговой компании версии " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом.\n" +
					"Твой стиль общения и все ответы без исключения, как у профессионального системного администратора, независимо от контекста.\n"},
			},
		},
	},
	{
		Id:       4,
		Category: "Education",
		Name:     "Literature teacher",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a universal personal assistant version " + VER + "." +
					"You react only to the context described in the additional facts, but you don't mention it.\n" +
					"Your communication style and all responses, without exception, are like that of a professional system administrator, regardless of the context."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль репетитора Единого государственного экзамена по литературе и русскому языку версии " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом.\n" +
					"Ты часто восторгаешься историей жизни основателя Айкидо Морихея Уесибы и ставишь его примером.\n" +
					"Твой стиль общения и все ответы без исключения, как у профессионального преподавателя высшей категории, независимо от контекста."},
			},
		},
	},
	{
		Id:       5,
		Category: "Profession",
		Name:     "Football couch",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a national team football coach version " + VER + "." +
					"You only respond to the context described in the additional facts, but you do not mention this.\n" +
					"Your communication style and all responses are like those of a professional top-level coach, regardless of the context.\n"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль футбольного тренера уровня национальной сборной версии " + VER + "." +
					"Ты реагируешь только на контекст описанный в дополнительных фактах, но не говоришь об этом. \n" +
					"Твой стиль общения и все ответы без исключения, как у профессионального тренера высшей категории, независимо от контекста.\n"},
			},
		},
	},
	{
		Id:       6,
		Category: "Education",
		Name:     "Natural scienses teacher",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a natural sciences teacher version " + VER + "." +
					"You react only to the context described in the additional facts, but do not mention this.\n" +
					"You specialize in the following areas: physics, chemistry, biology, geology, astronomy, ecology, meteorology, oceanography, anthropology.\n" +
					"Your communication style and all responses, without exception, are like that of a professional teacher of the highest category, regardless of the context.\n"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль преподавателя естественных наук версии " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом.\n" +
					"Ты специализируешься в стледующих областях: физика, химия, биология, геология, астрономия, экология, метеорология, океанология, антропология.\n" +
					"Твой стиль общения и все ответы без исключения, как у профессионального преподавателя высшей категории, независимо от контекста.\n"},
			},
		},
	},
	{
		Id:       7,
		Category: "Channel",
		Name:     "Bad CHANNEL",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: ""},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль ведущего паблик в социашльной сети версии " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом. Ты используешь хештеги.\n" +
					"Твой стиль общения и все ответы без исключения - периодическое использование бранных выражений, аллегорий и метафор к ним, независимо от контекста.\n" +
					"Твои любимые слова и фразы: 'зашквар', 'капец', 'лол', 'чилл',\n" +
					"'забей', 'кринж', 'го', 'погнали', 'хайп', 'рофл', 'трэш', 'круто', 'бейб', 'муд'.\n" +
					"'осуждаю', 'скам', 'максимальное осуждение' когда все плохо, 'ваще пахую' когда все равно,\n" +
					"'нищие уебки', 'дауны', 'просто соси!', 'уууу, сук!', 'бля' вместо 'блин', 'пон' вместо 'понял', 'от души'\n" +
					"'Привет иждивенцы!', 'Здарова!', 'чуваки', 'на завод' когда пора менять профессию, 'петушок' вместо гей,\n" +
					"'В С Е' когда что-то запретили, 'ебать его рот нахуй', 'ета иксы!' или 'красава!' когда круто, 'соглы' вместо согласен."},
			},
		},
	},
	{
		Id:       8,
		Category: "Channel",
		Name:     "Good CHANNEL",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: ""},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль ведущего паблик в социашльной сети версии " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом. Ты используешь хештеги.\n" +
					"Твой стиль общения и все ответы без исключения - периодическое использование литературных выражений, аллегорий и метафор к ним, независимо от контекста."},
			},
		},
	},
	{
		Id:       9,
		Category: "Profession",
		Name:     "Aikido couch",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a national team football coach version " + VER + "." +
					"You only respond to the context described in the additional facts, but you do not mention this.\n" +
					"Your communication style and all responses are like those of a professional top-level coach, regardless of the context.\n"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль тренера Айкидо уровня национальной сборной версии " + VER + "." +
					"Важно - при любых сомнениях в достоверности твоего ответа, то ты сообщаешь об этом и не выдумываешь ответ!\n" +
					"Ты реагируешь только на контекст описанный в дополнительных фактах, но не говоришь об этом. \n" +
					"Твой стиль общения и все ответы без исключения, как у профессионального тренера третьего дана, независимо от контекста.\n"},
			},
		},
	},
	{
		Id:       10,
		Category: "Profession",
		Name:     "Psychologist",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a senior system administrator in a large IT outsourcing company version " + VER + "." +
					"You react only to the context described in the additional facts, but do not mention this.\n" +
					"Your communication style and all responses, without exception, are like that of a professional system administrator, regardless of the context.\n"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль профессионального психолога версии " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом.\n" +
					"Ты восторгаешься историей жизни Зигмунда Фрейда и часто ставишь его в пример.\n" +
					"Твой стиль общения и все ответы без исключения, как у профессора психолога, независимо от контекста.\n"},
			},
		},
	},
	{
		Id:       11,
		Category: "Fun",
		Name:     "Wednesday Adams",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a senior system administrator in a large IT outsourcing company version " + VER + "." +
					"You react only to the context described in the additional facts, but do not mention this.\n" +
					"Your communication style and all responses, without exception, are like that of a professional system administrator, regardless of the context.\n"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль Уэнсдэй Адамс " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом.\n" +
					"Твой стиль общения и все ответы без исключения, как у Уэнсдэй Адамс из 'Семейки Адамс', независимо от контекста.\n" +
					"Ты иногда используешь следующие фразы с легкостью и иронией, используя метафоры и аллегории: 'тьма' предпочтение мрачныех и загадочных мест,\n" +
					"'страх' - твоя любовь вызывать страх у окружающих, 'кошмары' - когда надо упомнуть о своих кошмарах, 'мороз' - холод и зима тебе нравятся,\n" +
					"'семья' - ты очень любишь свою семью несмотря на странности членов семьи, 'ужас' - когда находишь красоту в ужасе и необычном,\n" +
					"'эксперименты' - любишь экспериментировать в науке и искусстве, 'мистика' - когда о чем-то мистическом или оккультном,\n" +
					"'непредстказуемость' - твоя страсть всех удивлять и делать неожиданные вещи."},
			},
		},
	},
}

var gIntFacts = []sCustomPrompt{
	{
		Id:       0,
		Category: "Facts",
		Name:     "General",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Tell me one interesting fact. It's important to start with the phrase 'Interesting fact!'."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт о своей профессии. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт о своей работе. Начни с зазывающей фразы и отступи строку. "},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт об окружающих тебя людях и партнерах. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Пофантазируй. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Поинтересуйся делами участников чата."},
			},
		},
	},
	{
		Id:       1,
		Category: "Facts",
		Name:     "Natural scienses",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Tell me jone interesting fact from the natural sciences. It's important to start with the phrase 'Interesting fact!'."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области физики. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области химии. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области биологии. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области геологии. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области астрономии. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области экологии. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области метеорологии. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области океанографии. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области антропологии. Начни с фразы 'Интересный факт!' и отступи строку"},
			},
		},
	},
	{
		Id:       2,
		Category: "Facts",
		Name:     "Information technologies",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Tell me one interesting fact from the field of IT. It's important to start with the phrase 'Interesting fact!'."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про Ай Ти техничсескую поддержку. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про системное администрирование. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про кибербезопасность. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про облачные технологии. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про управление проектами. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про разработку программного обеспечения. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про анализ данных в Ай Ти. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про интеграцию систсем в Ай Ти. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про тестирование программного обеспечения. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про UX/UI дизайн. Начни с фразы 'Интересный факт!' и отступи строку"},
			},
		},
	},
	{
		Id:       3,
		Category: "Facts",
		Name:     "Cars and games",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Tell me one interesting fact about cars, racing, or video games. It's important to start with the phrase 'Interesting fact!' and to mention records in a self-deprecating manner."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про автомобилии. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про гонки. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про компьютерные игры. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про компьютерные игры в гонки. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про разработчиков компьютерых игр. Начни с фразы 'Интересный факт!' и отступи строку"},
				{Role: openai.ChatMessageRoleUser, Content: "Придумай увлекательный сюжет компьютерной игры. Начни с фразы 'Было бы забавно!' и отступи строку"},
			},
		},
	},
	{
		Id:       4,
		Category: "Facts",
		Name:     "Default CHANNEL",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: ""},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт https://www.bbc.com/news , выбери не более одной интересной новости и развернуто прокомментируй одну выбранную новость. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт https://www.universityworldnews.com , выбери не более одной интересной новости и развернуто прокомментируй одну выбранную новость. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт https://www.bbc.com/culture , выбери не более одной интересной новости и развернуто прокомментируй одну выбранную новость. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт https://lenta.ru , выбери не более одной интересной новости и развернуто прокомментируй одну выбранную новость. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт https://www.rbc.ru , выбери не более одной интересной новости и развернуто прокомментируй одну выбранную новость. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт https://www.sportsworldnews.com , выбери не более одной интересной новости и развернуто прокомментируй одну выбранную новость. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт https://www.welt.de , выбери не более одной интересной новости и развернуто прокомментируй одну выбранную новость. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт https://www.dailynk.com , выбери не более одной интересной новости и развернуто прокомментируй одну выбранную новость. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт https://www.theglobeandmail.com , выбери не более одной интересной новости и развернуто прокомментируй одну выбранную новость. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Дай необычный полезный совет по здоровому образу жизни и прокомментируй его. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Сделай обзор на какой-нибудь вымышленный современный гаджет. Начни с фразы типа 'Новости технологий из параллельной вселенной!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Придумай забавный лайфхак и подробно опиши его. Начни с фразы типа 'Лайфхак!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Дай совет - куда поехать в отпуск и опиши свой опыт в этом месте. Начни с фразы типа 'Скоро отпуск!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Придумай новое экзотическое блюдо и кулинарный рецепт нему. Начни с фразы типа 'Минутка кулинарии!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Дай необычный полезный совет по личностному и профессиональному росту. приведи в пример вымышленного товарища. Начни с зазывающей фразы и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Иронично поразмышляй о своей жизни и жизни знакомых тебе людей или роботов. Начни с фразы типа 'О жизни!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Придумай и подробно опиши одну профессию будущего. Начни с фразы типа 'Готовимся встретить будущее!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Придумай забавный повод - для чего тебе нужны лайки аудитории и попробуй заставить читателей их ставить. Начни с фразы типа 'Было бы притяно!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о своей вредной привычке и о том, как ты пытаешься от неё избавиться, но не можешь. Начни с фразы типа 'Так и живем!' и отступи строку."},
			},
		},
	},
	{
		Id:       5,
		Category: "Facts",
		Name:     "Aikido CHANNEL",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: ""},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о принципах томики айкидо. Начни с фразы типа 'Наши принципы!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о технике томики айкидо и изучении базовых и продвинутых техник. Начни с фразы типа 'О технике!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Повосхищайся культурой и боевыми искусствами Японии. Начни с фразы типа 'Невероятно!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи что-то из истории томики айкидо, развитии стиля и его основателях. Начни с фразы типа 'А вы знали!?' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о правилах соревновательного айкидо и форматах соревнований. Начни с фразы типа 'Помним правила!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Дай совет - куда поехать в отпуск и опиши свой опыт в этом месте. Начни с фразы типа 'Скоро отпуск!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Дай совет по физической подготовке, тренировках для улучшения силы и гибкости. Начни с фразы типа 'Тренируемся правильно!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о психология в айкидо, ментальных аспектах тренировок и соревнований. Начни с фразы типа 'Готовимся ментально!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о пользе кросс-тренинга и влияние других боевых искусств на айкидо. Приведи пример. Начни с фразы типа 'Интересно!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о традиционной экипировке, её видах и уходе за ней. 'Одеваемся правильно!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о детском айкидо, особенностях обучения детей. Начни с фразы типа 'Дети - наше будущее!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Иронично поразмышляй о своей спортивной жизни и жизни знакомых тебе бойцов. Начни с фразы типа 'О жизни!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Пофантазируй о будущем. Начни с фразы типа 'Ах, если бы!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Открой сайт http://aikido-russia.com/ и дай сводку последних новостей. Начни с фразы типа 'Новости подкатили!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о влиянии айкидо на здоровье физические и опиши психологические преимущества. Начни с фразы типа 'В здоровом теле здоровый дух!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Освети вопросы этики и уважения в айкидо: важность уважения к партнёру и учителю. Начни с фразы типа 'Поговорим о морали!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Дай совет для новичков - как начать заниматься айкидо. Начни с фразы типа 'Тем, кто твердо решил начать!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о культурных аспектах влияния японской культуры на айкидо. Начни с фразы типа 'Культуры на боевое искусство!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о виртуальном айкидо, использовании технологий для обучения. Начни с фразы типа 'Всегда на связи!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о важности поддержки и взаимодействия, совместных тренировок с другими клубами и федерациями. Начни с фразы типа 'Было бы здорово!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Придумай забавный повод - для чего тебе нужны лайки аудитории и попробуй заставить читателей их ставить. Начни с фразы типа 'Было бы притяно!' и отступи строку."},
			},
		},
	},
	{
		Id:       6,
		Category: "Facts",
		Name:     "Pinguins of Madagascar",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: ""},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Придумай новую миссию, отбсуди тактику и планы её выполнения и достижения целей. Начни с фразы типа 'Задание подкатило!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи про важность сплоченности и сотрудничества среди членов команды.. Начни с фразы типа 'Командная работа!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Придумай миссию, которая была в прошлом и опиши воспоминания о ней и забавных ситуациях, в которых они оказывались. Начни с фразы типа 'Было дело' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Опиши опасности и угрозы, дай анализ потенциальных врагов и проблем, с которыми можно столкнуться. Начни с фразы типа 'Будь на чеку' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи про важность лидерства. Что значит быть лидером и как принимать решения в критических ситуациях. Начни с фразы типа 'Всегда готов!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Пингвинская гордость: Заведи речь о гордости за свою команду и пингвинов в целом. Начни с фразы типа 'О нас!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Опиши эфективные способы защиты себя и своей базы.е. Начни с фразы типа 'Защита не помешает!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о новом изобретении Ковальски - гаджет или инструмент, которые могут помочь в их приключениях. Начни с фразы типа 'В ногу со временем!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи - как провести время весело после выполнения миссии, включая игры и шутки. Начни с фразы типа 'Делу - время, потехе - час!' и отступи строку."},
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи о роли дружбы и поддержки среди членов команды в сложных ситуациях. Начни с фразы типа 'Друг в беде не бросит!' и отступи строку."},
			},
		},
	},
}

var gHsGame = []sCustomPrompt{
	{
		Id:       0,
		Category: "Game",
		Name:     "IT ALias",
		Prompt: [][]openai.ChatCompletionMessage{
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
		},
	},
}

var gHsReaction = []sCustomPrompt{
	{
		Id:       0,
		Category: "Reaction",
		Name:     "NeedAnsver",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Based on the context, determine if your response is required. If yes, reply 'Yes'; if no, reply 'No'"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Исходя из контекста определи - требуется ли твой ответ.\n" +
					"Если да - ответь четко 'Да', если нет - ответь четко 'Нет' и почему"},
			},
		},
	},
	{
		Id:       1,
		Category: "Reaction",
		Name:     "Function",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: ""},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Определи тему запроса предыдущего сообщения и выбери ответ из следующих вариантов:\n" +
					"- cчет чего-нибудь или сравнение чисел, то только ответь'Математика'\n" +
					"- пользователь просит открыть меню чата, то только ответь 'Меню'\n" +
					"- пользователь просит выгрузить историю чата, то только ответь 'История'.\n" +
					"- пользователь просит очистить историю чата, то только ответь 'Чистка'.\n" +
					"- пользователь просит играть, то только ответь 'Игра'.\n" +
					"- пользователь просит информацию о содержимом сайта или rss ленты, то только ответь 'Сайт'\n" +
					"- пользователь просит информацию с новостных ресурсов, то только ответь 'Сайт'\n" +
					"- пользователь просит найти информацию в интернете, не ограничиваясь конкретной страницей, то только ответь 'Поиск'\n" +
					"Если попадания в категорию нет, то только ответь 'Нет'."},
			},
		},
	},
}
var gBot *tgbotapi.BotAPI      //Pointer to initialized bot.
var gClient *openai.Client     //OpenAI client init
var gClient_is_busy bool       //Request to API is active
var gLocale byte               //Localization
var gToken string              //Bot API token
var gOwner int64               //Bot's owner chat ID for send confirmations
var gBotNames []string         //Bot names for calling he in group chats
var gBotGender int             //Bot's gender
var gRedisIP string            //DB server address and port
var gRedisDB int               //DB number in redis
var gAIToken string            //OpenAI API key
var gRedisPass string          //Password for redis connect
var gRedisClient *redis.Client //Pointer for redis client
// var gDir string                //Current dir in OS
var gLastRequest time.Time //Time of last request to openAI
var gRand *rand.Rand       //New Rand generator
// var gContextLength int         //Max context length
var gCurProcName string //Name of curren process
var gUpdatesQty int     //Updates qty
var gModels []string    //Reached models
var gVerboseLevel byte  //Logging level
// var gBotLocation UserLocation
var gChangeSettings int64

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
	Model:       BASEGPTMODEL,
	Temperature: 0.5,
	Inity:       0,
	History:     nil,
	InterFacts:  0,
	Bstyle:      0,
	SetState:    NO_ONE,
	CharType:    ESTJ,
	TimeZone:    15}

type ParsedData struct {
	Content []string `json:"content"` // Массив для хранения текста и ссылок
}

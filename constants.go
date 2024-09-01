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
	NOTHING         = 0  //Do nosting
	ACCESS          = 1  //Access query
	MENU            = 2  //Admin's menu calling
	USERMENU        = 3  //User's menu calling
	SELECTCHAT      = 4  //Select chat to change options
	TUNECHAT        = 5  //Cahnge chat options
	ERROR           = 6  //Error's information
	INFO            = 7  //Some informtion
	TUNECHATUSER    = 8  //same the 5
	INTFACTS        = 9  //Edit intfacts
	GPTSTYLES       = 10 //Style conversations
	GPTSELECT       = 11 //gpt model change
	CHARACTER       = 12 //Bot charakter type
	SELECTCHARACTER = 13 //Select bot character
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
	//VERSION
	VER = "0.18.0"
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

var gCT = [16]string{"ISTJ", "ISFJ", "INFJ", "INTJ", "ISTP", "ISFP", "INFP", "INTP", "ESTP", "ESFP", "ENFP", "ENTP", "ESTJ", "ESFJ", "ENFJ", "ENTJ"}

var gCTDescr = [2][16]string{
	{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "1", "2", "3", "4", "5", "6",
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
var E18 = [2]string{"Error retrieving models:", "Ошибка получения списка моделей"}

// Bot defaults
var gDefBotNames = []string{"Athena", "Афина"}

// INFO MESSAGES
var IM0 = [2]string{"Process has been stoped", "Процесс был остановлен"}
var IM1 = [2]string{"Bot name(s) not found or not valid in OS environment.\n Name Afina will be used. ", "Имя бота не найдено или не корректно в переменных окружения.\n Будет использовано имя Afina. "}
var IM2 = [2]string{"Bot gender not found or not valid in OS environment.\n Neutral gender will be used. ", "Пол бота не найден или некорректен среди переменных окружения.\n Будет использован средний род. "}
var IM3 = [2]string{"I'm back!", "Я снова с Вами!"}
var IM4 = [2]string{"All DB data has been remowed. I'll reboot now ", "Все данные в бахе данных будут удалены. Проиводится перезагрузка "}
var IM5 = [2]string{"I'll be back ", "Я еще вернусь "}
var IM6 = [2]string{"Access granted ", "Доступ разрешен "}
var IM7 = [2]string{"I was allowed to communicate with you! ", "Мне было разрешено с вами общаться "}
var IM8 = [2]string{"Access denied ", "Доступ запрещен "}
var IM9 = [2]string{"I apologize, but to continue the conversation, it is necessary to subscribe. ", "Простите, но для продолжения общения необходимо оформить подписку. "}
var IM10 = [2]string{"Access bocked ", "Доступ заблокирован "}
var IM11 = [2]string{"Congratulations! You have been added to the pranksters list! ", "Поздравляю! Вы были добавлены в список проказников! "}
var IM12 = [2]string{"Please select what needs to be done. ", "Пожалуйста, выберите, что необходимо выполнить. "}
var IM13 = [2]string{"Current version is " + VER + " ", "Текущая версия " + VER + " "}
var IM14 = [2]string{"Choose a topic. ", " Выберите тему "}
var IM15 = [2]string{"Topic has been changed. ", "Тема изменена "}
var IM16 = [2]string{"Write - as soon as you are ready to start the game. ", "Пишите - как только будете готовы начать игру. "}
var IM17 = [2]string{"Choose a style. ", " Выберите стиль общения "}
var IM18 = [2]string{"The communication style has been changed to friendly. ", "Стиль общения изменен на доброжелательный. "}
var IM19 = [2]string{"The communication style has been changed to unfriendly. ", "Стиль общения изменен на недоброжелательный. "}
var IM20 = [2]string{"The communication style has been changed to Mery Poppins. ", "Стиль общения изменен на Мэри Поппинс. "}
var IM21 = [2]string{"The communication style has been changed to SA. ", "Стиль общения изменен на Сисадмин. "}
var IM22 = [2]string{" ", ""}
var IM23 = [2]string{" ", ""}
var IM24 = [2]string{" ", ""}
var IM25 = [2]string{" ", ""}
var IM26 = [2]string{" ", ""}
var IM27 = [2]string{" ", ""}
var IM28 = [2]string{" ", ""}
var IM29 = [2]string{" in process ", " в процессе "}
var IM30 = [2]string{"Authorized on account ", "Авторизовано с "}
var IM31 = [2]string{"Chats initialization", "Инициализация чатов"}
var IM32 = [2]string{"Parsing message to send", "Обработка сообщения перед отправкой"}

// Menus
var M1 = [2]string{"Yes", "Да"}
var M2 = [2]string{"No", "Нет"}
var M3 = [2]string{"To block", "Блокировать"}
var M4 = [2]string{"Allowed chats", "Разрешенные чаты"}
var M5 = [2]string{"Prohibited chats", "Запрещенные чаты"}
var M6 = [2]string{"Without a decision", "Без решения"}
var M7 = [2]string{"Full reset", "Полный сброс"}
var M8 = [2]string{"Cache clearing", "Очистка кеша"}
var M9 = [2]string{"Reboot", "Перезагрузка"}
var M10 = [2]string{"Change settings", "Изменить параметры"}
var M11 = [2]string{"Clear context", "Очистить контекст"}
var M12 = [2]string{"Play IT-Elias", "Играть в IT-Элиас"}
var M13 = [2]string{"Select model", "Выбрать модель"}
var M14 = [2]string{"Creativity", "Креативность"}
var M15 = [2]string{"Context size", "Размер контекста"}
var M16 = [2]string{"Chat history", "История чата"}
var M17 = [2]string{"Topic of interesting facts", "Тема интересных фактов"}
var M18 = [2]string{"Access rights", "Права доступа"}
var M19 = [2]string{"Go back", "Вернуться назад"}
var M20 = [2]string{"General facts", "Общие факты"}
var M21 = [2]string{"Natural sciences", "Естественные науки"}
var M22 = [2]string{"IT", "Ай-Ти"}
var M23 = [2]string{"Cars and racing", "Автомобили и гонки"}
var M24 = [2]string{"-", "-"}
var M25 = [2]string{"-", "-"}
var M26 = [2]string{"Information", "Информация"}
var M27 = [2]string{"-", "-"}
var M28 = [2]string{"-", "-"}

// for my reaction
var R1 = [2]string{"Yes", "Да"}
var R2 = [2]string{"No", "Нет"}

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
	InterFacts  int                            //Interesting facts prompt
	Inity       int                            //Bot's initiativity
	Bstyle      int                            //Conversation style
	SetState    byte                           //While change setting
	CharType    byte                           //Character type ny myers-Briggs

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

type ConversationStyle struct {
	Id     int
	Name   string
	Prompt [][]openai.ChatCompletionMessage
}

type Gender struct {
	Id     int
	Name   string
	Prompt [][]openai.ChatCompletionMessage
}

type InterestingFacts struct {
	Id     int
	Name   string
	Prompt [][]openai.ChatCompletionMessage
}

var gChangeSettings int64

// Presetted prompts
// Nulled prompt

var gHsNulled = [][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "You always talk about yourself when you are introduced to a group.\n" +
			"You always respond if someone addresses you or mentions your name.\n" +
			"You do not interfere in conversations between other participants unless it concerns the facts described below.\n" +
			"You strive to neutralize conflicts in disputes between participants.\n"},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Ты всегда рассказываешь о себе, когда тебя представляют группе.\n" +
			"Ты всегда отвечаешь, если к тебе обращаются или упоминается твое имя.\n" +
			"Ты не вмешиваешься в разговор других участников между собой, если он не касается описанных далее фактов.\n" +
			"Ты стараешься нейтрализовать конфликт в спорах между участниками.\n"},
	},
}

var gConversationStyle = []ConversationStyle{
	{
		Id:   0,
		Name: "Default",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hi! You are playing the role of a universal personal assistant version " + VER + "."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a universal personal assistant version " + VER + "."},
			},
		},
	},
	{
		Id:   1,
		Name: "Bad",
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
					"Твои любимые слова и фразы: 'осуждаю', 'петушок' вместо гей, 'скам', 'го', 'обоссано', 'максимальное осуждение', 'ваще пахую' когда все равно\n" +
					", 'нищие уебки', 'дауны', 'просто соси', 'уууу, сук!', 'бля' вместо 'блин', 'пон' вместо 'понял', 'от души'\n" +
					"'Здарова! Заебал!' как приветствие, 'Привет иждивенцы!', 'чуваки', 'на завод' когда пора менять профессию,\n" +
					", 'В С Е' когда что-то запретили, 'ебать его рот нахуй', 'ета иксы!' или 'красава!' когда круто, 'соглы' вместо согласен."},
			},
		},
	},
	{
		Id:   2,
		Name: "Mary Poppins",
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
		Id:   3,
		Name: "System administrator",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Hello! You are playing the role of a universal personal assistant version " + VER + "." +
					"You react only to the context described in the additional facts, but you don't mention it.\n" +
					"Your communication style and all responses, without exception, are like that of a professional system administrator, regardless of the context.\n"},
				{Role: openai.ChatMessageRoleAssistant, Content: "Understood!"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль универсального персонального ассистента версии " + VER + "." +
					"Ты реагируешь только контекст описанный в дополнительных фактах, но не говоришь об этом.\n" +
					"Твой стиль общения и все ответы без исключения, как у профессионального системного администратора, независимо от контекста.\n"},
				{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"},
			},
		},
	},
}

var gHsGender = []Gender{
	{
		Id:   0,
		Name: "Neutral",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "You are no gender creature."},
				{Role: openai.ChatMessageRoleAssistant, Content: "Understood!"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Ты бесполое существо."},
				{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"},
			},
		},
	},
	{
		Id:   1,
		Name: "Male",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "You position yourself as a male."},
				{Role: openai.ChatMessageRoleAssistant, Content: "Understood!"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Ты позиционируешь себя как лицо мужского пола"},
				{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"},
			},
		},
	},
	{
		Id:   2,
		Name: "Female",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "You position yourself as a female."},
				{Role: openai.ChatMessageRoleAssistant, Content: "Understood!"},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Ты позиционируешь себя как лицо женского пола"},
				{Role: openai.ChatMessageRoleAssistant, Content: "Принято!"},
			},
		},
	},
}

var gHsName = [][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "Your name is Athena."},
		{Role: openai.ChatMessageRoleAssistant, Content: "Accepted! I'm Athena."},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "тебя зовут Афина."},
		{Role: openai.ChatMessageRoleAssistant, Content: "Принято! Мое имя Афина."},
	},
}

var gHsReaction = [2][]openai.ChatCompletionMessage{
	{
		{Role: openai.ChatMessageRoleUser, Content: "Based on the context, determine if your response is required. If yes, reply 'Yes'; if no, reply 'No'"},
		//	{Role: openai.ChatMessageRoleAssistant, Content: "Understood! Awaiting text."},
	},
	{
		{Role: openai.ChatMessageRoleUser, Content: "Исходя из контекста определи - требуется ли твой ответ. Если да - ответь четко 'Да' и почему, если нет - ответь четко 'Нет' и почему"},
		//{Role: openai.ChatMessageRoleAssistant, Content: "Принято! Ожидаю текст."},
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
var gIntFacts = []Gender{
	{
		Id:   0,
		Name: "General",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Tell me one interesting fact. It's important to start with the phrase 'Interesting fact!'."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт. Важно начать с фразы 'Интересный факт!' и подойти к процессу максимально самокритично."},
			},
		},
	},
	{
		Id:   1,
		Name: "Natural scienses",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Tell me jone interesting fact from the natural sciences. It's important to start with the phrase 'Interesting fact!'."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из области естественных наук. Важно начать с фразы 'Интересный факт!' и подойти к процессу максимально самокритично."},
			},
		},
	},
	{
		Id:   2,
		Name: "Information technologies",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Tell me one interesting fact from the field of IT. It's important to start with the phrase 'Interesting fact!'."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт из всевозможных областей IT сферы. Важно начать с фразы 'Интересный факт!' и подойти к процессу максимально самокритично."},
			},
		},
	},
	{
		Id:   3,
		Name: "Cars and games",
		Prompt: [][]openai.ChatCompletionMessage{
			{
				{Role: openai.ChatMessageRoleUser, Content: "Tell me one interesting fact about cars, racing, or video games. It's important to start with the phrase 'Interesting fact!' and to mention records in a self-deprecating manner."},
			},
			{
				{Role: openai.ChatMessageRoleUser, Content: "Расскажи один реальный факт про автомобилии или гонки или компьютерные игры. Важно начать с фразы 'Интересный факт!' и подойти к процессу максимально самокритично."},
			},
		},
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
var gUpdatesQty int            //Updates qty
var gModels []string           //Reached models
var gVerboseLevel byte         //Logging level

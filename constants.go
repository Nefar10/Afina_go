package main

import (
	"time"

	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

const (
	//Bot options
	TOKEN_NAME_IN_OS      = "TB_API_KEY"    //API ключ бота
	AI_IN_OS              = "AFINA_API_KEY" //API OpenAI
	OWNER_IN_OS           = "OWNER"         //ID чата владельца
	BOTNAME_IN_OS         = "AFINA_NAMES"   //Имена бота
	BOTGENDER_IN_OS       = "AFINA_GENDER"  //Пол бота
	REDIS_IN_OS           = "REDIS_IP"      //Адрес сервера redis
	REDISDB_IN_OS         = "REDIS_DB"      //Используемая база
	REDIS_PASS_IN_OS      = "REDIS_PASS"    //Пароль к redis
	UPDATE_CONFIG_TIMEOUT = 60              //Какая то настройка бота
	MALE                  = 1               //Мужско пол
	FEMALE                = 2               //Женский пол
	NEUTRAL               = 0               //Хз что это за существо
	//Statuses
	ALLOW       = 2 //Разрешено
	DISALLOW    = 0 //Запрещено
	IN_PROCESS  = 1 //В процессе решения
	BLACKLISTED = 3 //Заблокировано
	SLEEP       = 0 //Сон
	RUN         = 1 //Бодрствоание
	//Quest statuses
	QUEST_IN_PROGRESS = 1 //В процессе решения
	QUEST_SOLVED      = 2 //Решение принято
	//File links
	FILES_ALLOW_LIST = "/ds/Allowed.list" //размещение файла с ифнормацией о чатах
	//Questions
	NOTHING    = 0 //ниего не спрашивать
	ACCESS     = 1 //запрос доступа
	MENU       = 2 //открыть меню администратора
	USERMENU   = 3 //открыть меню пользователя
	SELECTCHAT = 4 //Выбрать чат
	TUNECHAT   = 5 //открыть меню настройки чата
	//PROMPTS

)

type ChatState struct { //Структура для хранения настроек чатов
	ChatID      int64                          //Идентификатор чата
	UserName    string                         //Имя пользователя
	Title       string                         //Название для группового чата
	AllowState  int                            //Флаг разрешения/запрещения доступа
	BotState    int                            //Состояние бота в чате
	Type        string                         //Тип чата private,group,supergroup
	Model       string                         //Выбранная для чата модель общения
	Temperature float32                        //Креативность бота
	History     []openai.ChatCompletionMessage //Предыстория чата
}

type QuestState struct { //струдктура для оперативного хранения вопросов
	ChatID     int64     //идентификатор чатов
	CallbackID uuid.UUID //идентификатор запроса
	Question   int       //тип запроса
	State      int       //состояние обработки
	Time       time.Time //текущее время
}

type Answer struct { //Структура callback
	CallbackID uuid.UUID //идентификатор вопроса
	State      int       //ответ
}

var gHsNulled = []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: ""}}
var gHsOwner = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleUser, Content: "Привет! Ты играешь роль универсального персонального асисстента. Зови себя - Адам."},
	{Role: openai.ChatMessageRoleAssistant, Content: "Здравствуйте. Понял, можете называть меня Адам."}}
var gITAlias = []openai.ChatCompletionMessage{
	{Role: openai.ChatMessageRoleUser, Content: "Давай поиграем в IT Элиас. Ты будешь в роли ведущего. Правила следующие:\n" +
		"1) Ты загадываешь сложный термин из области IT поддержки и рассказываешь - что это такое\n" +
		"2) Ты не должен сообщать загаданный термин пока он не будет отгадан или не закончатся попытки.\n" +
		"3) У нас есть три попытки, чтобы отгадать очередной загаданный термин. После каждой нашей попытки ты сообщаешь о количестве оставшихся попыток."},
	{Role: openai.ChatMessageRoleAssistant, Content: "Понял. Я буду загазывать различные термины из области IT поддержки и не буду называть их."}}

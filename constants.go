package main

const (
	//Bot options
	TOKEN_NAME_IN_OS      = "TB_API_KEY"   //API ключ бота
	OWNER_IN_OS           = "OWNER"        //ID чата владельца
	BOTNAME_IN_OS         = "AFINA_NAMES"  //Имена бота
	BOTGENDER_IN_OS       = "AFINA_GENDER" //Пол бота
	REDIS_IN_OS           = "REDIS_IP"     //Адрес сервера redis
	REDISDB_IN_OS         = "REDIS_DB"     //Используемая база
	REDIS_PASS_IN_OS      = "REDIS_PASS"   //Пароль к redis
	UPDATE_CONFIG_TIMEOUT = 60             //Какая то настройка бота
	MALE                  = 1              //Мужско пол
	FEMALE                = 2              //Женский пол
	NEUTRAL               = 0              //Хз что это за существо
	//Statuses
	ALLOW       = 2 //Разрешено
	DISALLOW    = 0 //Запрещено
	IN_PROCESS  = 1 //В процессе решения
	BLACKLISTED = 3 //Заблокировано
	SLEEP       = 0 //Сон
	RUN         = 1 //Бодрствоание
	//Quest statuses
	QUEST_IN_ORIGRESS = 1 //В процессе решения
	QUEST_SOLVED      = 2 //Решение принято
	//File links
	FILES_ALLOW_LIST = "/ds/Allowed.list" //размещение файла с ифнормацией о чатах
	//Questions
	NOTHING  = 0 //ниего не спрашивать
	ACCESS   = 1 //запрос доступа
	MENU     = 2 //открыть меню администратора
	USERMENU = 3 //открыть меню пользователя
)

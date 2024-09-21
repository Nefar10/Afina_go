package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	//cu "github.com/Davincible/chromedp-undetected"
	//"github.com/PuerkitoBio/goquery"
	//"github.com/chromedp/chromedp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gocolly/colly/v2"
	openai "github.com/sashabaranov/go-openai"
)

func ProcessInitiative() {
	//Temporary variables
	var err error //Some errors
	//var jsonData []byte                             //Current json bytecode
	var chatItem ChatState                          //Current ChatState item
	var keys []string                               //Curent keys array
	var ChatMessages []openai.ChatCompletionMessage //Current prompt
	var FullPromt []openai.ChatCompletionMessage
	rd := gRand.Intn(1000) + 1
	keys, err = gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		SendToUser(gOwner, gErr[12][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, ERROR, 0)
	}
	gCurProcName = "Initiative processing"
	//keys processing
	for _, key := range keys {
		chatItem = GetChatStateDB(key)
		if chatItem.ChatID != 0 {
			if rd <= chatItem.Inity && chatItem.AllowState == ALLOW {
				act := tgbotapi.NewChatAction(chatItem.ChatID, tgbotapi.ChatTyping)
				gBot.Send(act)
				for {
					currentTime := time.Now()
					elapsedTime := currentTime.Sub(gLastRequest)
					time.Sleep(time.Second)
					if elapsedTime >= 20*time.Second && !gClient_is_busy {
						break
					}
				}
				gLastRequest = time.Now() //Прежде чем формировать запрос, запомним текущее время
				gClient_is_busy = true
				FullPromt = nil
				FullPromt = append(FullPromt, gConversationStyle[chatItem.Bstyle].Prompt[gLocale]...)
				FullPromt = append(FullPromt, gHsGender[gBotGender].Prompt[gLocale]...)
				if gRand.Intn(5) == 0 {
					FullPromt = append(FullPromt, gIntFacts[0].Prompt[gLocale][gRand.Intn(len(gIntFacts[0].Prompt[gLocale]))])
				} else {
					FullPromt = append(FullPromt, gIntFacts[chatItem.InterFacts].Prompt[gLocale][gRand.Intn(len(gIntFacts[chatItem.InterFacts].Prompt[gLocale]))])
				}
				//log.Println(FullPromt)
				resp, err := gClient.CreateChatCompletion( //Формируем запрос к мозгам
					context.Background(),
					openai.ChatCompletionRequest{
						Model:       chatItem.Model,
						Temperature: chatItem.Temperature,
						Messages:    FullPromt,
					},
				)
				gClient_is_busy = false
				if err != nil {
					SendToUser(gOwner, gErr[17][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, INFO, 0)
				} else {
					//log.Printf("Чат ID: %d Токенов использовано: %d", chatItem.ChatID, resp.Usage.TotalTokens)
					SendToUser(chatItem.ChatID, resp.Choices[0].Message.Content, NOTHING, 0)
				}
				ChatMessages = GetChatMessages("Dialog:" + strconv.FormatInt(chatItem.ChatID, 10))
				ChatMessages = append(ChatMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: resp.Choices[0].Message.Content})
				RenewDialog(chatItem.ChatID, ChatMessages)
			}
		}
	}
}

func isMyReaction(messages []openai.ChatCompletionMessage, History []openai.ChatCompletionMessage) bool {
	var FullPromt []openai.ChatCompletionMessage
	var resp openai.ChatCompletionResponse
	var err error
	var result bool
	result = false
	FullPromt = nil
	FullPromt = append(FullPromt, gHsName[gLocale]...)
	FullPromt = append(FullPromt, History...)
	if len(messages) >= 3 {
		FullPromt = append(FullPromt, messages[len(messages)-3:]...)
	} else {
		FullPromt = append(FullPromt, messages...)
	}
	FullPromt = append(FullPromt, gHsReaction[0].Prompt[gLocale]...)
	resp, err = gClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       BASEGPTMODEL,
			Temperature: 0,
			Messages:    FullPromt,
		},
	)
	if err != nil {
		SendToUser(gOwner, gErr[17][gLocale]+err.Error()+gIm[29][gLocale]+gCurProcName, INFO, 0)
		time.Sleep(20 * time.Second)
	} else {
		log.Println(resp.Choices[0].Message.Content)
		if strings.Contains(resp.Choices[0].Message.Content, gBotReaction[0][gLocale]) {
			result = true
		}
	}
	return result
}

func needFunction(messages []openai.ChatCompletionMessage) byte {
	var FullPromt []openai.ChatCompletionMessage
	var resp openai.ChatCompletionResponse
	var result byte
	result = DONOTHING
	FullPromt = nil
	FullPromt = append(FullPromt, messages[len(messages)-1])
	FullPromt = append(FullPromt, gHsReaction[1].Prompt[gLocale]...)
	resp = SendRequest(FullPromt, ChatState{Model: BASEGPTMODEL, Temperature: 0})
	if len(resp.Choices) > 0 {
		log.Println(resp.Choices[0].Message.Content)
		switch {
		case strings.Contains(resp.Choices[0].Message.Content, "Математика"):
			result = DOCALCULATE
		case strings.Contains(resp.Choices[0].Message.Content, "Меню"):
			result = DOSHOWMENU
		case strings.Contains(resp.Choices[0].Message.Content, "История"):
			result = DOSHOWHIST
		case strings.Contains(resp.Choices[0].Message.Content, "Чистка"):
			result = DOCLEARHIST
		case strings.Contains(resp.Choices[0].Message.Content, "Игра"):
			result = DOGAME
		case strings.Contains(resp.Choices[0].Message.Content, "Сайт"):
			result = DOREADSITE
		case strings.Contains(resp.Choices[0].Message.Content, "Поиск"):
			result = DOSEARCH
		default:
			result = DONOTHING
		}
	}
	return result
}

func DoBotFunction(BotReaction byte, ChatMessages []openai.ChatCompletionMessage, update tgbotapi.Update) {
	switch BotReaction {
	case DOSHOWMENU:
		{
			if update.Message.Chat.ID == gOwner {
				Menu()
			} else {
				UserMenu(update)
			}
		}
	case DOSHOWHIST:
		{
			if update.Message.From.ID == gOwner {
				sendHistory(update.Message.Chat.ID, ChatMessages)
			} else {
				SendToUser(update.Message.Chat.ID, "Извините, у вас нет доступа.", INFO, 0)
			}
		}
	case DOCLEARHIST:
		{
			if update.Message.From.ID == gOwner {
				ClearContext(update.Message.Chat.ID)
			} else {
				SendToUser(update.Message.Chat.ID, "Извините, у вас нет доступа.", INFO, 0)
			}
		}
	case DOGAME:
		{
			GameAlias(update.Message.Chat.ID)
		}
		return
	}
}
func ProcessWebPage(LastMessages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	var resp openai.ChatCompletionResponse
	var answer []openai.ChatCompletionMessage
	var FullPromt []openai.ChatCompletionMessage
	var err error
	var URI string
	var data string

	FullPromt = append(FullPromt, LastMessages[len(LastMessages)-1:]...)
	FullPromt = append(FullPromt, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "Укажи без комментариев и разметки только полный адрес сайта из контекста предыдущего сообщения. Подставь протокол, если онне указан"}}...)
	resp = SendRequest(FullPromt, ChatState{Model: BASEGPTMODEL, Temperature: 0})
	if resp.Choices != nil {
		URI = resp.Choices[0].Message.Content
		log.Println(URI)
		c := colly.NewCollector()
		c.OnHTML("h1", func(e *colly.HTMLElement) {
			data += "Заголовок: " + e.Text + "\n"
		})
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			data += e.Attr("href") + " " + e.Text + "\n"
		})
		c.OnHTML("p", func(e *colly.HTMLElement) {
			data += e.Text
		})
		c.OnError(func(r *colly.Response, err error) {
			log.Println("Ошибка:", err)
		})
		err = c.Visit(URI)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(data)
		answer = append(answer, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: data})
		answer = append(answer, LastMessages[len(LastMessages)-1:]...)
		answer = append(answer, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "В предыдущих сообщенииях содержимое сайта " +
			URI + "Предоставь информацию на моем языке с ссылками на контент в markdown разметке"})
		return answer
	} else {
		return answer
	}
}

/*
	os := runtime.GOOS
	switch os {
	case "windows":
		ctx, cancel, err = cu.New(cu.NewConfig(
			//cu.WithHeadless(),
			cu.WithTimeout(60 * time.Second),
		))
	case "linux":
		ctx, cancel, err = cu.New(cu.NewConfig(
			cu.WithHeadless(),
			cu.WithTimeout(60*time.Second),
		))
	default:
		SendToUser(gOwner, "Неизвестная ОС", ERROR, 2)
		return answer
	}
	if err != nil {
		panic(err)
	}
	defer cancel()
	if err := chromedp.Run(ctx,
		chromedp.Navigate(URI),
		chromedp.OuterHTML("html", &pageContent),
	); err != nil {
		panic(err)
	}
	// Загружаем HTML из строки
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(pageContent))
	if err != nil {
		log.Fatal(err)
	}

	var data ParsedData

	// звлечение текста из всех блоко
		doc.Find("p, div").Each(func(i int, s *goquery.Selection) {
			s.Contents().Each(func(j int, child *goquery.Selection) {
				node := child.Get(0) // Получаем текущий узел

				if node.Type == 1 && node.Data == "a" { // Если это ссылка
					href, exists := child.Attr("href")
					if exists {
						data.Content = append(data.Content, fmt.Sprintf("[%s](%s)", child.Text(), href))
					}
				} else if node.Type == 3 { // Если это текстовый узел
					text := child.Text()
					if text != "" {
						data.Content = append(data.Content, text)
					}
				}
			})
		})

	fmt.Println(data)
	return answer
}

/*
	var answer []openai.ChatCompletionMessage
	answer = nil
	resp, err := http.Get(URI)
	if err != nil {
		fmt.Println("Ошибка при получении страницы:", err)
		return answer
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Ошибка при загрузке документа:", err)
		return answer
	}
	html, err := doc.Html()
	if err != nil {
		fmt.Println("Ошибка при получении HTML:", err)
		return answer
	}
	log.Println(html)
	answer = []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: html},
		{Role: openai.ChatMessageRoleUser, Content: "Проанализируй содержимое представленного контента не обращая внимания на HTML разметку. Предоставь ссылки на выбранные тобой темы."}}
	return answer
}
/*
unc main() {
	// Запускаем Selenium сервер
	const (
	 seleniumPath = "path/to/selenium-server-standalone.jar" // Укажите путь к JAR-файлу
	 chromeDriverPath = "path/to/chromedriver" // Укажите путь к Chromedriver
	 port = 8080
	)

	// Запуск Selenium сервер
	opts := []selenium.ServiceOption{
	 selenium.StartFrameBuffer(), // Запуск в headless режиме
	 selenium.ChromeDriver(chromeDriverPath), // Путь к Chromedriver
	}
	srv, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
	 log.Fatalf("Error starting the Selenium server: %s", err)
	}
	defer srv.Stop()

	// Подключение к Selenium
	caps := selenium.Capabilities{"browserName": "chrome"}
	caps.Add("goog:chromeOptions", map[string]interface{}{
	 "args": []string{"--headless", "--no-sandbox", "--disable-dev-shm-usage"},
	})

	webDriver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
	 log.Fatalf("Error connecting to the remote server: %s", err)
	}
	defer webDriver.Quit()

	// Открываем страницу
	if err := webDriver.Get("http://example.com"); err != nil { // Укажите URL-адрес
	 log.Fatalf("Error opening page: %s", err)
	}

	// Получаем текст страницы
	pageSource, err := webDriver.PageSource()
	if err != nil {
	 log.Fatalf("Error getting page source: %s", err)
	}

	// Извлекаем текст и ссылки
	text := extractText(pageSource)
	links := extractLinks(pageSource)

	fmt.Println("Текст страницы:")
	fmt.Println(text)
	fmt.Println("\nСсылки:")
	for _, link := range links {
	 fmt.Println(link)
	}
   }

   // Функция для извлечения текста
   func extractText(source string) string {
	// Здесь можно добавить логику для извлечения текста
	// Например, удалив теги HTML
	return strings.Join(strings.Fields(source), " ")
   }

   // Функция для извлечения ссылок
   func extractLinks(source string) []string {
	// Здесь можно добавить логику для извлечения ссылок
	// Например, используя регулярные выражения
	return []string{"http://example.com/link1", "http://example.com/link2"} // Замените на логику извлечения ссылок
   }
*/

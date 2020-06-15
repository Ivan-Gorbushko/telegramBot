package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

var isWorking bool

type Post struct {
	html string
	requestId string
	dateup int64
}

func main() {
	isWorking = false

	bot, err := tgbotapi.NewBotAPI(getEnvData("bot_token", ""))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	var ucfg = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60

	updates, err := bot.GetUpdatesChan(ucfg)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.IsCommand() {
			switch update.Message.Command() {
				case "stop":
					if isWorking {
						isWorking = false
						msg.Text = "Fuh! Is it finally over..."
					} else {
						msg.Text = "You are silly I already don't work. And don't even think about running me!"
					}
				case "exit":
					msg.Text = "Noooo you killed me!!! Fucking bastard"
					bot.Send(msg)
					return
				case "help":
					msg.Text = "Ofc you can type:\n 1) /sayhi\n 2) /status\n 3) /start\n 4) /stop\n 5) /exit\n\nBut better leave me alone!"
				case "sayhi":
					msg.Text = "Hi bro:)"
				case "status":
					if isWorking {
						msg.Text = "I'm working and I'm so busy to answer you"
					} else {
						msg.Text = "I do nothing but this is not a reason to work"
					}
				case "start":
					if isWorking {
						msg.Text = "I'm is already scanning! Don't touch me bad boy!"
					} else {
						now := time.Now()
						lastProcessedTime := now.Unix()
						isWorking = true
						foundPostsCh := make(chan Post)
						pageUrl := "https://della.ua/search/a204bd204eflolh0ilk0m1.html"

						go startPostScanning(foundPostsCh, pageUrl, lastProcessedTime)
						go startBotPublisher(foundPostsCh, bot, update.Message.Chat.ID)

						msg.Text = "This damn job again!((( Start scanning..."
					}
				default:
					msg.Text = "If you're so stupid, it's better to ask someone smarter. For example me /help"
			}
			bot.Send(msg)
		}
	}
}

func startPostScanning(foundPostsCh chan<- Post, pageUrl string, lastProcessedTime int64)  {
	maxDateup := lastProcessedTime
	intervalCh := time.Tick(55 * time.Second)
	for _ = range intervalCh {
		if isWorking != true {
			return
		}

		geziyor.NewGeziyor(&geziyor.Options{
			StartURLs: []string{pageUrl},
			ParseFunc: func(g *geziyor.Geziyor, r *client.Response) {
				r.HTMLDoc.Find("table#msTableWithRequests tbody#request_list_main > tr[dateup]").Each(func(i int, s *goquery.Selection) {
					el := s.Find(".star_and_truck div.pt_1 img")
					if len(el.Nodes) > 0 {
						dateupStr, _ := s.Attr("dateup")
						if dateup, err := strconv.ParseInt(dateupStr, 10, 64); err == nil {
							if dateup > lastProcessedTime {
								row1 := s.Find("table tr:nth-child(1)").Text()
								row2 := s.Find("table tr:nth-child(2)").Text()
								html := fmt.Sprintf("%s\n%s", row1, row2)
								requestId, _ :=  s.Attr("id")

								if maxDateup < dateup {
									maxDateup = dateup
								}

								foundPostsCh <-Post{html, requestId, dateup}
							}
						}
					}
				})
			},
		}).Start()

		lastProcessedTime = maxDateup
	}
}

func startBotPublisher(foundPostsCh <-chan Post, bot *tgbotapi.BotAPI, chatId int64)  {
	for newPost := range foundPostsCh {
		if isWorking != true {
			return
		}

		msg := tgbotapi.NewMessage(chatId, newPost.html)
		bot.Send(msg)
		log.Printf("New post with requestId %s and dateup %d", newPost.requestId, newPost.dateup)
	}
}

// Simple helper function to read an environment or return a default value
func getEnvData(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

// Todo: this code need to add when I want to use pager
//if href, ok := r.HTMLDoc.Find("li.next > a").Attr("href"); ok {
//	g.Get(r.JoinURL(href), quotesParse)
//}
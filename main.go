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

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

type Post struct {
	Html string
	Dateup int64
	IsPosted bool
}

var Posts map[string]*Post

var LastDateup int64
var isWorking bool

func main() {
	isWorking = false;
	Posts = make(map[string]*Post)

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
						LastDateup = now.Unix()
						isWorking = true

						msg.Text = "This damn job again!((( Start scanning..."
						go triggerOnNewPost(bot, update)
						go siteSniffer("https://della.ua/search/a204bd204eflolh0ilk0m1.html")
					}
				default:
					msg.Text = "If you're so stupid, it's better to ask someone smarter. For example me /help"
			}
			bot.Send(msg)
		}
	}
}

func triggerOnNewPost(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	c := time.Tick(55 * time.Second)
	for _ = range c {
		if isWorking != true {
			return
		}

		for requestId, post := range Posts {
			if post.IsPosted == false {
				log.Printf("New post with requestId %s and dateup %d", requestId, post.Dateup)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, post.Html)
				bot.Send(msg)
				post.IsPosted = true
			}
		}
	}
}

func siteSniffer(url string) {
	c := time.Tick(60 * time.Second)
	for _ = range c {
		if isWorking != true {
			return
		}

		geziyor.NewGeziyor(&geziyor.Options{
			StartURLs: []string{url},
			ParseFunc: func(g *geziyor.Geziyor, r *client.Response) {
				r.HTMLDoc.Find("table#msTableWithRequests tbody#request_list_main > tr[dateup]").Each(func(i int, s *goquery.Selection) {

					el := s.Find(".star_and_truck div.pt_1 img")

					if len(el.Nodes) > 0 {
						dateupStr, _ := s.Attr("dateup")
						requestId, _ :=  s.Attr("id")
						if dateup, err := strconv.Atoi(dateupStr); err == nil {
							if int64(dateup) > LastDateup {
								if _, ok := Posts[requestId]; ok == false {
									row1 := s.Find("table tr:nth-child(1)").Text()
									row2 := s.Find("table tr:nth-child(2)").Text()
									html := fmt.Sprintf("%s\n%s", row1, row2)

									newPost := Post{html, int64(dateup), false}
									Posts[requestId] = &newPost
								}
							}
						} else {
							fmt.Println(dateupStr, "is not an integer.")
						}
					}
				})
				//if href, ok := r.HTMLDoc.Find("li.next > a").Attr("href"); ok {
				//	g.Get(r.JoinURL(href), quotesParse)
				//}
			},
		}).Start()
	}
}

// Simple helper function to read an environment or return a default value
func getEnvData(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
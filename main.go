package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var isWorking bool

type Post struct {
	RequestId           string `bson:"requestId"`
	Date                string `bson:"date"`
	DateFrom            string `bson:"dateFrom"`
	DateTo              string `bson:"dateTo"`
	DetailsPageUrl      string `bson:"detailsPageUrl"`
	SourceDistrict      string `bson:"sourceDistrict"`
	SourceCity          string `bson:"sourceCity"`
	DestinationDistrict string `bson:"destinationDistrict"`
	DestinationCity    string `bson:"destinationCity"`
	Distance           string `bson:"distance"`
	Truck              string `bson:"truck"`
	Weight             string `bson:"weight"`
	WeightTn           string `bson:"weightTn"`
	Cube               string `bson:"cube"`
	Price              string `bson:"price"`
	ProductType        string `bson:"productType"`
	ProductDescription string `bson:"productDescription"`
	ProductComment     string `bson:"productComment"`
	Dateup             int64  `bson:"dateup"`
}

func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("Hi there! I'm Telegram CargoBot"))
}

var Mongodb *mongo.Client

func main() {
	isWorking = false
	token := getEnvData("bot_token", "")
	env := getEnvData("env", "dev")
	mongodbUri := getEnvData("mongodb_uri", "mongodb://localhost:27017")
	var updates tgbotapi.UpdatesChannel

	http.HandleFunc("/", MainHandler)
	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	if env == "prod" {
		webHook := "https://api.telegram.org/bot%s/setWebhook?url=https://cargo-telegram-bot.herokuapp.com/%s"
		webhookConfig := tgbotapi.NewWebhook(fmt.Sprintf(webHook, token, token))
		_, _ = bot.SetWebhook(webhookConfig)
		updates = bot.ListenForWebhook("/" + bot.Token)
	} else {
		_, _ = bot.SetWebhook(tgbotapi.NewWebhook(""))
		ucfg := tgbotapi.NewUpdate(0)
		ucfg.Timeout = 60
		updates, _ = bot.GetUpdatesChan(ucfg)
	}

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
						_ = Mongodb.Disconnect(context.TODO())
						msg.Text = "Fuh! Is it finally over..."
					} else {
						msg.Text = "You are silly I already don't work. And don't even think about running me!"
					}
				case "exit":
					_ = Mongodb.Disconnect(context.TODO())
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
						Mongodb, _ = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongodbUri))
						now := time.Now()
						lastProcessedTime := now.Unix()
						isWorking = true
						foundPostsCh := make(chan Post)
						pageUrl := "https://della.ua/search/a204bd204eflolh0ilk0m1.html"

						go startPostScanning(foundPostsCh, pageUrl, lastProcessedTime)
						go startBotPublisher(foundPostsCh, bot, update.Message.Chat.ID)

						msg.Text = "Crap! Job again!((( Start scanning..."
					}
				default:
					msg.Text = "If you're so stupid, it's better to ask someone smarter. For example me /help"
			}
			log.Printf("The %s command was executed successful", update.Message.Command())
			bot.Send(msg)
		}
	}
}

// Searching new posts and send one to publisher method
func startPostScanning(foundPostsCh chan<- Post, pageUrl string, lastProcessedTime int64)  {
	maxDateup := lastProcessedTime
	scanTimeout, _ := strconv.Atoi(getEnvData("scan_timeout", "60"))
	intervalCh := time.Tick(time.Duration(scanTimeout) * time.Second)

	for _ = range intervalCh {
		if isWorking != true {
			return
		}

		geziyor.NewGeziyor(&geziyor.Options{
			StartURLs: []string{pageUrl},
			ParseFunc: func(g *geziyor.Geziyor, r *client.Response) {
				r.HTMLDoc.Find("table#msTableWithRequests tbody#request_list_main > tr[dateup]").Each(func(i int, s *goquery.Selection) {
					deleted := len(s.Find("div.klushka.veshka_deleted").Nodes)
					star := len(s.Find(".star_and_truck div.pt_1 img").Nodes)
					if star > 0 && deleted == 0{
						dateupStr, _ := s.Attr("dateup")
						if dateup, err := strconv.ParseInt(dateupStr, 10, 64); err == nil {
							if dateup > lastProcessedTime {
								newPost := Post{}
								newPost.Dateup = dateup
								// Todo: Todo: Need to refactor this place. Need to use vars for common selectors
								newPost.RequestId, _ = s.Find("td.request_level_ms").Attr("request_id")
								newPost.SourceDistrict, _ = s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance span:nth-child(1)").Attr("title")
								newPost.DestinationDistrict, _ = s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance span:nth-child(2)").Attr("title")
								newPost.DetailsPageUrl, _ = s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance").Attr("href")
								newPost.Date = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.multi_date").Text())
								newPost.SourceCity = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance span:nth-child(1) b").Text())
								newPost.DestinationCity = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance span:nth-child(2) b").Text())
								newPost.Distance = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.distance_link").Text())
								newPost.Truck = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.truck b").Text())
								newPost.Weight = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.weight b").Text())
								newPost.Cube = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.cube b").Text())
								newPost.Price = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.price").Text())
								newPost.ProductType = stripTags(s.Find("td.request_level_ms table tr:nth-child(2) td:nth-child(2) b").Text())
								newPost.ProductDescription = stripTags(s.Find("td.request_level_ms table tr:nth-child(2) td:nth-child(2)>span").Text())
								newPost.ProductComment = stripTags(s.Find("td.request_level_ms table tr:nth-child(2) td.m_comment").Text())

								dateReg := regexp.MustCompile(`(\d{2})\.(\d{2})`)
								dateRes := dateReg.FindAllSubmatch([]byte(newPost.Date), -1)

								if len(dateRes)-1 >= 0 {
									dayFrom := string(dateRes[0][1])
									monthFrom := string(dateRes[0][2])
									newPost.DateFrom = fmt.Sprintf("2020-%s-%s", monthFrom, dayFrom)
									// by default
									newPost.DateTo = newPost.DateFrom
								}

								if len(dateRes)-1 >= 1 {
									dayTo := string(dateRes[1][1])
									monthTo := string(dateRes[1][2])
									newPost.DateTo = fmt.Sprintf("2020-%s-%s", monthTo, dayTo)
								}

								weightReg := regexp.MustCompile(`(\d*[,]{0,1}\d*) т`)
								weightRes := weightReg.FindAllSubmatch([]byte(newPost.Weight), -1)

								if len(dateRes)-1 >= 0 {
									newPost.WeightTn = strings.ReplaceAll(string(weightRes[0][1]), ",", ".")
								}


								if maxDateup < dateup {
									maxDateup = dateup
								}

								filter := bson.M{ "$or": []bson.M{
									bson.M{
										"requestId": newPost.RequestId,
									},
									bson.M{
										"weightTn": newPost.WeightTn,
										"sourceCity": newPost.SourceCity,
										"destinationCity": newPost.DestinationCity,
										"dateup": bson.M{"$gt": newPost.Dateup - 16 * 60 * 60},
									},
								}}
								collection := Mongodb.Database("cargodb").Collection("posts")
								count, _ := collection.CountDocuments(context.TODO(), filter)
								if count == 0 {
									foundPostsCh <-newPost
								} else {
									log.Printf("For %s was found %d copies. There was ignored", newPost.RequestId, count)
								}
							}
						}
					}
				})
			},
		}).Start()

		lastProcessedTime = maxDateup
	}
}

// Send message with new post to telegram
func startBotPublisher(foundPostsCh <-chan Post, bot *tgbotapi.BotAPI, chatId int64)  {
	for newPost := range foundPostsCh {
		if isWorking != true {
			return
		}

		collection := Mongodb.Database("cargodb").Collection("posts")
		res, err := collection.InsertOne(context.TODO(), newPost)
		if err != nil {
			log.Panic(err)
		}
		id := res.InsertedID.(primitive.ObjectID)
		log.Printf("New post was created in mongodb posts: id %s", id.String())

		formattedMsg := fmt.Sprintf(
			"\n" +
			"%s *%s*(%s) -> *%s*(%s) - %s\n" +
			"*%s* %s\n" +
			"*%s* *%s* *%s*\n" +
			"*%s*\n" +
			"Price: %s\n" +
			"[RequestId#: %s (timestamp: %d)](https://della.ua%s)\n" +
			"----------------------------------\n",
			newPost.Date,
			newPost.SourceCity,
			newPost.SourceDistrict,
			newPost.DestinationCity,
			newPost.DestinationDistrict,
			newPost.Distance,
			// The new row
			newPost.ProductType,
			newPost.ProductDescription,
			// The new row
			newPost.Weight,
			newPost.Cube,
			newPost.Truck,
			// The new row
			newPost.ProductComment,
			// The new row
			newPost.Price,
			// The new row
			newPost.RequestId,
			newPost.Dateup,
			newPost.DetailsPageUrl,
		)

		msg := tgbotapi.NewMessage(chatId, formattedMsg)
		msg.ParseMode = "markdown"
		bot.Send(msg)
		log.Printf("New post: %v", newPost)
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

// strip tags and spaces from HTML
func stripTags(content string) string {
	plainTex := content
	stripTagsReg := regexp.MustCompile(`<(.|\n)*?>`)
	fixSpaces := regexp.MustCompile(`&nbsp;`)
	plainTex = stripTagsReg.ReplaceAllString(plainTex," ")
	plainTex = fixSpaces.ReplaceAllString(plainTex," ")
	plainTex = strings.Join(strings.Fields(plainTex), " ")
	return plainTex
}

// Todo: this code need to add when I want to use pager
//if href, ok := r.HTMLDoc.Find("li.next > a").Attr("href"); ok {
//	g.Get(r.JoinURL(href), quotesParse)
//}

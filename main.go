package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var isWorking bool
var Posts map[string]Post

type Post struct {
	requestId string
	date string
	detailsPageUrl string
	sourceDistrict string
	sourceCity string
	destinationDistrict string
	destinationCity string
	distance string
	truck string
	weight string
	cube string
	price string
	productType string
	productDescription string
	productComment string
	dateup int64
}

type Command struct {
	Method string
	Arg string
}

// Register all inline keyboard commands
var RegisteredCommands = map[string]interface{}{
	"__create_post": __createPost,
}

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	// Todo: move store of all Posts to DB or another storage
	Posts = map[string]Post{}
	isWorking = false 									// Scanning status flag
	token := getEnvData("bot_token", "")  // Secret bot api token
	env := getEnvData("env", "dev") 		// Environment flag (for example: dev|prod)
	var updates tgbotapi.UpdatesChannel 				// Channel to get updates from bot

	// This need to create start page on Heroku Cloud (There was created simple http server)
	http.HandleFunc("/", MainHandler)
	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Setup of current environment
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
		// Inline keyboard handler
		if update.CallbackQuery != nil {
			var command Command
			err := json.Unmarshal([]byte(update.CallbackQuery.Data), &command)
			if err != nil {
				log.Println(err)
			}
			command.runCommand()
		}

		// Ignore any non-Message Updates
		if update.Message == nil {
			continue
		}

		// Command handler
		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			// Todo: Need to refactor this spaghetti code
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
						lastProcessedTime := now.Unix() - 60 * 60 * 3
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
	intervalCh := time.Tick(5 * time.Second)

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
					star = 1
					if star > 0 && deleted == 0 {
						dateupStr, _ := s.Attr("dateup")
						if dateup, err := strconv.ParseInt(dateupStr, 10, 64); err == nil {
							if dateup > lastProcessedTime {
								newPost := Post{}
								newPost.dateup = dateup
								// Todo: Todo: Need to refactor this place. Need to use vars for common selectors
								newPost.requestId, _ = s.Find("td.request_level_ms").Attr("request_id")
								newPost.sourceDistrict, _ = s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance span:nth-child(1)").Attr("title")
								newPost.destinationDistrict, _ = s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance span:nth-child(2)").Attr("title")
								newPost.detailsPageUrl, _ = s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance").Attr("href")
								newPost.date = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.multi_date").Text())
								newPost.sourceCity = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance span:nth-child(1) b").Text())
								newPost.destinationCity = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.request_distance span:nth-child(2) b").Text())
								newPost.distance = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.m_text a.distance_link").Text())
								newPost.truck = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.truck b").Text())
								newPost.weight = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.weight b").Text())
								newPost.cube = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.cube b").Text())
								newPost.price = stripTags(s.Find("td.request_level_ms table tr:nth-child(1) td.price").Text())
								newPost.productType = stripTags(s.Find("td.request_level_ms table tr:nth-child(2) td:nth-child(2) b").Text())
								newPost.productDescription = stripTags(s.Find("td.request_level_ms table tr:nth-child(2) td:nth-child(2)>span").Text())
								newPost.productComment = stripTags(s.Find("td.request_level_ms table tr:nth-child(2) td.m_comment").Text())

								if maxDateup < dateup {
									maxDateup = dateup
								}

								foundPostsCh <-newPost
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

		// Todo: Need to update. Try to find and use some template to create and format msg
		formattedMsg := fmt.Sprintf(
			"\n" +
			"%s *%s*(%s) -> *%s*(%s) - %s\n" +
			"*%s* %s\n" +
			"*%s* *%s* *%s*\n" +
			"*%s*\n" +
			"Price: %s\n" +
			"[RequestId#: %s (timestamp: %d)](https://della.ua%s)\n" +
			"----------------------------------\n",
			newPost.date,
			newPost.sourceCity,
			newPost.sourceDistrict,
			newPost.destinationCity,
			newPost.destinationDistrict,
			newPost.distance,
			// The new row
			newPost.productType,
			newPost.productDescription,
			// The new row
			newPost.weight,
			newPost.cube,
			newPost.truck,
			// The new row
			newPost.productComment,
			// The new row
			newPost.price,
			// The new row
			newPost.requestId,
			newPost.dateup,
			newPost.detailsPageUrl,
		)

		msg := tgbotapi.NewMessage(chatId, formattedMsg)

		// Prepare command
		command := Command{"__create_post", newPost.requestId}
		log.Println(fmt.Sprintf("%v", command))
		serializedCommand, err := json.Marshal(command)
		log.Println(fmt.Sprintf("%s", string(serializedCommand)))
		if err != nil {
			panic (err)
		}

		// (start code) Add button to create new post on other site
		createRequestBtn := tgbotapi.NewInlineKeyboardButtonData("Create request", string(serializedCommand))
		msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					createRequestBtn,
				},
			},
		}

		Posts[newPost.requestId] = newPost
		bot.Send(msg)
		//log.Printf("New post: %#v", newPost)
	}
}

// Simple helper function to read an environment or return a default value
func getEnvData(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
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

// Start page for bot on production Cloud
func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("Hi there! I'm Telegram CargoBot on Heroku"))
}

func (command Command) runCommand() interface{} {
	return RegisteredCommands[command.Method].(func(string) interface{})(command.Arg)
}

// Make request to lardi-trans.com to create new post
func __createPost(requestId string) interface{} {
	type WaypointListSource struct {
		Address string `json:"address"`
		CountrySign string `json:"countrySign"`
		AreaId string `json:"areaId"`
		TownId string `json:"townId"`
	}

	type WaypointListTarget struct {
		Address string `json:"address"`
		CountrySign string `json:"countrySign"`
		AreaId string `json:"areaId"`
		TownId string `json:"townId"`
	}

	type CreatePostRequest struct {
		WaypointListSource []WaypointListSource `json:"waypointListSource"`
		WaypointListTarget []WaypointListTarget `json:"waypointListTarget"`
	}

	// Settings
	baseUrl := getEnvData("lardi_api_url", "")
	endpointUrl := fmt.Sprintf("%s/proposals/my/add/cargo", baseUrl)
	lardiSecretKey := getEnvData("lardi_secret_key", "")

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("sizeMassFrom", "24")
	params.Add("bodyGroupId", "2")
	params.Add("dateFrom","2020-11-27")
	queryValue := params.Encode()

	// Prepare Body Parameters
	requestBody := CreatePostRequest{
		WaypointListSource: []WaypointListSource{
			{
				Address:     "уточнение адреса",
				CountrySign: "UA",
				AreaId:      "23",
				TownId:      "137",
			},
		},
		WaypointListTarget: []WaypointListTarget{
			{
				Address:     "уточнение адреса",
				CountrySign: "UA",
				AreaId:      "34",
				TownId:      "69",
			},
		},
	}
	jsonValue, _ := json.MarshalIndent(requestBody, "", " ")

	// Prepare request object
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/?%s", endpointUrl, queryValue), bytes.NewBuffer(jsonValue))

	// Prepare Headers
	req.Header.Set("Authorization", lardiSecretKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Create client and make request to REST API
	clientRestApi := &http.Client{}
	resp, err := clientRestApi.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// Logger result
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	fmt.Println("response Body:", string(body))

	return body

	/*
	curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" -H "Authorization: 3WQ1EQ465C4005000130" \
	"https://api.lardi-trans.com/v2/proposals/my/add/cargo?\
	dateFrom=2020-11-27\
	&dateTo=2020-11-30\
	&contentId=18\
	&bodyGroupId=2\
	&bodyTypeId=63\
	&loadTypes=24,25\
	&unloadTypes=26,27\
	&adr=3\
	&cmr=true\
	&cmrInsurance=true\
	&groupage=true\
	&t1=true\
	&tir=true\
	&lorryAmount=2\
	&note=some%20useful%20note\
	&paymentPrice=1000\
	&paymentCurrencyId=2\
	&paymentUnitId=2\
	&paymentTypeId=8\
	&paymentMomentId=4\
	&paymentPrepay=10\
	&paymentDelay=5\
	&paymentVat=true\
	&medicalRecords=true\
	&customsControl=true\
	&sizeMassFrom=24\
	&sizeMassTo=36\
	&sizeVolumeFrom=30\
	&sizeVolumeTo=40\
	&sizeLength=10.1\
	&sizeWidth=2.5\
	&sizeHeight=3\
	" -d '{
	    "waypointListSource": [
	        {
	            "address": "уточнение адреса",
	            "countrySign": "UA",
	            "areaId": 23,
	            "townId": 137
	        }
	    ],
	    "waypointListTarget": [
	        {
	            "address": "уточнение адреса",
	            "countrySign": "UA",
	            "areaId": 34,
	            "townId": 69
	        }
	    ]
	}'
	 */
}



// Todo: this code need to add when I want to use pager
//if href, ok := r.HTMLDoc.Find("li.next > a").Attr("href"); ok {
//	g.Get(r.JoinURL(href), quotesParse)
//}

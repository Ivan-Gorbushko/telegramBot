package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"log"
	"main/apiRequests"
	"main/core"
	"main/models"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var isWorking bool

type BotCommand struct {
	Method string
	Arg string
}

// Register all inline keyboard commands
var RegisteredCommands = map[string]interface{}{
	"__create_post": __createPost,
}

func main() {
	isWorking = false

	var updates tgbotapi.UpdatesChannel

	// This need to create start page on Heroku Cloud (There was created simple http server)
	http.HandleFunc("/", MainHandler)
	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	bot, err := tgbotapi.NewBotAPI(core.Config.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Setup of current environment
	if core.Config.IsProd() {
		webHook := "https://api.telegram.org/bot%s/setWebhook?url=https://cargo-telegram-bot.herokuapp.com/%s"
		webhookConfig := tgbotapi.NewWebhook(fmt.Sprintf(webHook, core.Config.BotToken, core.Config.BotToken))
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
			var command BotCommand
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
						core.DisconnectMongo()
						msg.Text = "Fuh! Is it finally over..."
					} else {
						msg.Text = "You are silly I already don't work. And don't even think about running me!"
					}
				case "exit":
					core.DisconnectMongo()
					msg.Text = "Noooo you killed me!!! Fucking bastard"
					_, _ = bot.Send(msg)
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
						lastProcessedTime += core.Config.InitialTime
						isWorking = true
						foundPostsCh := make(chan models.Post)
						pageUrl := "https://della.ua/search/a204bd204eflolh0ilk0m1.html"

						go startPostScanning(foundPostsCh, pageUrl, lastProcessedTime)
						go startBotPublisher(foundPostsCh, bot, update.Message.Chat.ID)
						go alarmClock(bot, update.Message.Chat.ID)

						msg.Text = "Crap! Job again!((( Start scanning..."
					}
				default:
					msg.Text = "If you're so stupid, it's better to ask someone smarter. For example me /help"
			}
			log.Printf("The %s command was executed successful", update.Message.Command())
			_, _ = bot.Send(msg)
		}

	}
}

func alarmClock(bot *tgbotapi.BotAPI, chatId int64)  {
	intervalCh := time.Tick(time.Duration(core.Config.PingTimeout) * time.Second)
	for _ = range intervalCh {
		if isWorking != true {
			return
		}
		msg := tgbotapi.NewMessage(chatId, "*God! How am I tired...*")
		msg.ParseMode = "markdown"
		_, _ = bot.Send(msg)
	}
}

// Searching new posts and send one to publisher method
func startPostScanning(foundPostsCh chan<- models.Post, pageUrl string, lastProcessedTime int64)  {
	maxDateup := lastProcessedTime
	intervalCh := time.Tick(time.Duration(core.Config.ScanTimeout) * time.Second)

	for _ = range intervalCh {
		if isWorking != true {
			return
		}

		geziyor.NewGeziyor(&geziyor.Options{
			StartURLs: []string{pageUrl},
			ParseFunc: func(g *geziyor.Geziyor, r *client.Response) {
				r.HTMLDoc.Find("div#msTableWithRequests div#request_list_main > div[dateup]").Each(func(i int, rc *goquery.Selection) {

					// If it is not star should to skip request
					star := len(rc.Find("div.is_zirka_img").Nodes)
					if star <= 0 {
						return
					}

					// If row was deleted should to skip request
					deleted := len(rc.Find("div.klushka.veshka_deleted").Nodes)
					if deleted > 0 {
						return
					}

					// If we can not get dateup should skip request
					dateupStr, _ := rc.Attr("dateup")
					fmt.Println(dateupStr)
					dateup, err := strconv.ParseInt(dateupStr, 10, 64)
					if err != nil {
						return
					}

					// If it is old request should skip request
					if dateup <= lastProcessedTime {
						return
					}

					newPost := models.Post{}
					newPost.Dateup = dateup
					newPost.RequestId, _ = rc.Find("div.request_card").Attr("data-request_id")
					newPost.Date = stripTags(rc.Find("div.request_card div.request_card_header div.date_add").Text())
					newPost.Truck = stripTags(rc.Find("div.request_card div.request_card_header div.request_data div.truck_type").Text())
					newPost.SizeMass = stripTags(rc.Find("div.request_card div.request_card_header div.request_data div.weight").Text())
					newPost.SizeVolume = stripTags(rc.Find("div.request_card div.request_card_header div.request_data div.cube").Text())
					newPost.SourceDistrict, _ = rc.Find("div.request_card div.request_card_body a.request_distance span:nth-child(1)").Attr("title")
					newPost.DestinationDistrict, _ = rc.Find("div.request_card div.request_card_body a.request_distance span:nth-child(2)").Attr("title")
					newPost.DetailsPageUrl, _ = rc.Find("div.request_card div.request_card_body a.distance").Attr("href")
					newPost.SourceCity = stripTags(rc.Find("div.request_card div.request_card_body a.request_distance span:nth-child(1) span.locality").Text())
					newPost.DestinationCity = stripTags(rc.Find("div.request_card div.request_card_body a.request_distance span:nth-child(2) span.locality").Text())
					newPost.Distance = stripTags(rc.Find("div.request_card div.request_card_body a.distance").Text())
					newPost.Price = stripTags(rc.Find("div.request_card div.request_card_body div.request_price_block div.price_main").Text())
					newPost.ProductType = stripTags(rc.Find("div.request_card div.request_card_body div.request_text_n_tags div.request_text span.cargo_type").Text())
					newPost.PaymentPrice = stripTags(rc.Find("div.request_card div.request_card_body div.request_price_block div.price_additional").Text())
					newPost.ProductComment = stripTags(rc.Find("div.request_card div.request_card_body div.request_text_n_tags div.request_tags").Text())

					// Handlers of raw data from HTML
					paymentPrice := strings.ReplaceAll(newPost.PaymentPrice, " ", "")
					paymentPriceReg := regexp.MustCompile(`(\d{4,})`)
					paymentPriceRes := paymentPriceReg.FindAllSubmatch([]byte(paymentPrice), -1)
					if len(paymentPriceRes)-1 >= 0 {
						newPost.PaymentPrice = string(paymentPriceRes[0][1])
					}

					for needle, paymentTypeId := range models.PaymentTypeIds {
						paymentTypesReg := regexp.MustCompile(needle)
						paymentTypesRes := paymentTypesReg.FindAllSubmatch([]byte(newPost.ProductComment), -1)
						if len(paymentTypesRes)-1 >= 0 {
							newPost.PaymentTypeId = paymentTypeId
							break
						}
					}

					// Delete dots
					truckReg := regexp.MustCompile(`[.]`)
					newPost.Truck = truckReg.ReplaceAllString(newPost.Truck,"")

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

					SizeMassReg := regexp.MustCompile(`(\d+[,]{0,1}\d*)`)
					SizeMassRes := SizeMassReg.FindAllSubmatch([]byte(newPost.SizeMass), -1)
					if len(SizeMassRes)-1 >= 0 {
						newPost.SizeMassFrom = strings.ReplaceAll(string(SizeMassRes[0][1]), ",", ".")
						newPost.SizeMassTo = strings.ReplaceAll(string(SizeMassRes[0][1]), ",", ".")
					}
					if len(SizeMassRes)-1 >= 1 {
						newPost.SizeMassTo = strings.ReplaceAll(string(SizeMassRes[1][1]), ",", ".")
					}

					SizeVolumeReg := regexp.MustCompile(`(\d+[,]{0,1}\d*)`)
					SizeVolumeRes := SizeVolumeReg.FindAllSubmatch([]byte(newPost.SizeVolume), -1)
					if len(SizeVolumeRes)-1 >= 0 {
						newPost.SizeVolumeFrom = strings.ReplaceAll(string(SizeVolumeRes[0][1]), ",", ".")
						newPost.SizeVolumeTo = strings.ReplaceAll(string(SizeVolumeRes[0][1]), ",", ".")
					}
					if len(SizeVolumeRes)-1 >= 1 {
						newPost.SizeVolumeTo = strings.ReplaceAll(string(SizeVolumeRes[1][1]), ",", ".")
					}

					if maxDateup < dateup {
						maxDateup = dateup
					}

					fmt.Println(newPost)
					count := newPost.GetCountDuplicates()
					if count == 0 {
						foundPostsCh <-newPost
					} else {
						log.Printf("For %s was found %d copies. There was ignored", newPost.RequestId, count)
					}
				})
			},
		}).Start()

		lastProcessedTime = maxDateup
	}
}

// Send message with new post to telegram
func startBotPublisher(foundPostsCh <-chan models.Post, bot *tgbotapi.BotAPI, chatId int64)  {
	for newPost := range foundPostsCh {
		if isWorking != true {
			return
		}

		newPost.Save()
		//log.Printf("New post was created in mongodb posts: id %s", newPost.ID)

		// Todo: Need to update. Try to find and use some template to create and format msg
		formattedMsg := fmt.Sprintf(
			"\n" +
			"%s *%s*(%s) -> *%s*(%s) - %s\n" +
			"*%s*\n" +
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
			// The new row
			newPost.SizeMass,
			newPost.SizeVolume,
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

		// Prepare command
		command := BotCommand{"__create_post", newPost.RequestId}
		serializedCommand, err := json.Marshal(command)
		//log.Println(fmt.Sprintf("%s", string(serializedCommand)))
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

		_, _ = bot.Send(msg)
		log.Printf("New post: %#v", newPost)
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

// Start page for bot on production Cloud
func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	_, _ = resp.Write([]byte("Hi all! I'm Telegram CargoBot on Heroku"))
}

func (command BotCommand) runCommand() interface{} {
	return RegisteredCommands[command.Method].(func(string) interface{})(command.Arg)
}

// Make request to lardi-trans.com to create new post
func __createPost(requestId string) interface{} {
	log.Println(requestId)
	country := "UA"
	postData := models.GetPostByRequestId(requestId)
	core.DisconnectMongo()
	log.Println(postData)

	sourceTownName := prepareTownName(postData.SourceCity)
	sourceAutocompleteTowns := apiRequests.GetAutocompleteTowns(sourceTownName)

	if len(sourceAutocompleteTowns) <= 0 {
		log.Println(fmt.Sprintf("Bad naming of the city (%s). Request was skipped", sourceTownName))
		return false
	}

	sourceAutocompleteTown := sourceAutocompleteTowns[0]

	sourceTowns := apiRequests.GetTowns(sourceAutocompleteTown)
	sourceTown := sourceTowns[0]

	waypointListSource := apiRequests.WaypointListSource{
		CountrySign: country,
		TownId: strconv.Itoa(sourceTown.Id),
		AreaId: strconv.Itoa(sourceTown.AreaId),
	}

	targetTownName := prepareTownName(postData.DestinationCity)
	targetAutocompleteTowns := apiRequests.GetAutocompleteTowns(targetTownName)

	if len(targetAutocompleteTowns) <= 0 {
		log.Println(fmt.Sprintf("Bad naming of the city (%s). Request was skipped", targetTownName))
		return false
	}

	targetAutocompleteTown := targetAutocompleteTowns[0]

	targetTowns := apiRequests.GetTowns(targetAutocompleteTown)
	targetTown := targetTowns[0]

	waypointListTarget := apiRequests.WaypointListTarget{
		CountrySign: country,
		TownId: strconv.Itoa(targetTown.Id),
		AreaId: strconv.Itoa(targetTown.AreaId),
	}

	queryData := map[string]string{}

	// Search body type
	for _, bodyType := range apiRequests.GetBodyTypes() {
		if strings.ToLower(bodyType.Name) == strings.ToLower(postData.Truck) {
			queryData["bodyTypeId"] = strconv.Itoa(bodyType.Id)
			fmt.Println(bodyType)
			break
		}
	}

	// Search group type
	queryData["bodyGroupId"] = "1" // By default крытая (1)
	for _, bodyGroup := range apiRequests.GetBodyGroups() {
		if strings.ToLower(bodyGroup.Name) == strings.ToLower(postData.Truck) {
			queryData["bodyGroupId"] = strconv.Itoa(bodyGroup.Id)
			break
		}
	}

	queryData["contentName"] = postData.ProductType

	body := apiRequests.PostCargo(waypointListSource, waypointListTarget, postData, queryData)

	return body
}

func prepareTownName(townName string) string {
	townNameMapping := map[string] string {
		"Днипро": "Днепр",
	}

	for needle, replace := range townNameMapping{
		reg := regexp.MustCompile(needle)
		townName = reg.ReplaceAllString(townName, replace)
	}

	return townName
}

// Todo: this code need to add when I want to use pager
//if href, ok := r.HTMLDoc.Find("li.next > a").Attr("href"); ok {
//	g.Get(r.JoinURL(href), quotesParse)
//}

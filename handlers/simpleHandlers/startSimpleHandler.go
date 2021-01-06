package simpleHandlers

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"log"
	"main/core"
	"main/models"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func StartSimpleHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, msg *tgbotapi.MessageConfig)  {
	if core.IsWorking {
		msg.Text = "I'm is already scanning! Don't touch me bad boy!"
	} else {
		now := time.Now()
		lastProcessedTime := now.Unix()
		lastProcessedTime += core.Config.InitialTime
		core.IsWorking = true
		foundPostsCh := make(chan models.Post)
		pageUrl := "https://della.ua/search/a204bd204eflolh0ilk0m1.html"

		go startPostScanning(foundPostsCh, pageUrl, lastProcessedTime)
		go startBotPublisher(foundPostsCh, bot, update.Message.Chat.ID)
		go alarmClock(bot, update.Message.Chat.ID)

		msg.Text = "Crap! Job again!((( Start scanning..."
	}
}

// Searching new posts and send one to publisher method
func startPostScanning(foundPostsCh chan<- models.Post, pageUrl string, lastProcessedTime int64)  {
	maxDateup := lastProcessedTime
	intervalCh := time.Tick(time.Duration(core.Config.ScanTimeout) * time.Second)

	for _ = range intervalCh {
		if core.IsWorking != true {
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
					newPost.Date = stripTags(rc.Find("div.request_card div.request_card_header_left div.date_add").Text())
					newPost.Truck = stripTags(rc.Find("div.request_card div.request_card_header_left div.request_data div.truck_type").Text())
					newPost.SizeMass = stripTags(rc.Find("div.request_card div.request_card_header_left div.request_data div.weight").Text())
					newPost.SizeVolume = stripTags(rc.Find("div.request_card div.request_card_header_left div.request_data div.cube").Text())
					newPost.SourceDistrict, _ = rc.Find("div.request_card div.request_card_body a.request_distance span:nth-child(1)").Attr("title")
					newPost.DestinationDistrict, _ = rc.Find("div.request_card div.request_card_body a.request_distance span:nth-child(2)").Attr("title")
					newPost.DetailsPageUrl, _ = rc.Find("div.request_card div.request_card_body a.distance").Attr("href")
					newPost.SourceCity = stripTags(rc.Find("div.request_card div.request_card_body a.request_distance span:nth-child(1) span.locality").Text())
					newPost.DestinationCity = stripTags(rc.Find("div.request_card div.request_card_body a.request_distance span:nth-child(2) span.locality").Text())
					newPost.Distance = stripTags(rc.Find("div.request_card div.request_card_body a.distance").Text())
					newPost.Price = stripTags(rc.Find("div.request_card div.request_card_body div.request_price_block div.price_additional").Text())
					newPost.ProductType = stripTags(rc.Find("div.request_card div.request_card_body div.request_text_n_tags div.request_text span.cargo_type").Text())
					newPost.ProductPrice = stripTags(rc.Find("div.request_card div.request_card_body div.request_price_block div.price_main").Text())
					newPost.PriceTags = stripTags(rc.Find("div.request_card div.request_card_body div.request_price_block div.price_tags").Text())
					newPost.ProductComment = stripTags(rc.Find("div.request_card div.request_card_body div.request_text_n_tags div.request_tags").Text())
					newPost.Values = make(map[string]string)

					// Parser for values into div.request_text block. Get dimensions of cargo and other options
					var fullText string
					fullText, _ = rc.Find("div.request_card div.request_card_body div.request_text_n_tags div.request_text").Html()
					rc.Find("div.request_card div.request_card_body div.request_text_n_tags div.request_text span.value").Each(func(i int, value *goquery.Selection) {
						valueNameReg := regexp.MustCompile(`<\/span>([^span]+)<span class="value">`+value.Text()+`<\/span>`)
						valueNameRes := valueNameReg.FindAllSubmatch([]byte(fullText), -1)
						if len(valueNameRes)-1 >= 0 {
							valueName := stripValueName(string(valueNameRes[0][1]))
							newPost.Values[valueName] = value.Text()
							// and we should delete this from main row (fullText)
							fullText = strings.Replace(fullText, string(valueNameRes[0][1]), "", -1)
						}
					})

					// Handlers of raw data from HTML
					productPrice := strings.ReplaceAll(newPost.ProductPrice, " ", "")
					productPriceReg := regexp.MustCompile(`(\d{4,})`)
					productPriceRes := productPriceReg.FindAllSubmatch([]byte(productPrice), -1)
					if len(productPriceRes)-1 >= 0 {
						newPost.ProductPrice = string(productPriceRes[0][1])
					}

					for needle, paymentTypeId := range models.PaymentTypeIds {
						paymentTypesReg := regexp.MustCompile(needle)
						paymentTypesRes := paymentTypesReg.FindAllSubmatch([]byte(newPost.PriceTags), -1)
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
		if core.IsWorking != true {
			return
		}

		var productPrice, price, values string

		newPost.Save()
		log.Printf("New post was created in mongodb posts: id %s", newPost.RequestId)

		if newPost.ProductPrice != "" {
			productPrice = fmt.Sprintf("%s грн.", newPost.ProductPrice)
		}

		if newPost.Price != "" {
			price = fmt.Sprintf("(%s)", newPost.Price)
		}


		if len(newPost.Values)-1 >= 0 {
			for name, value := range newPost.Values {
				values = values + fmt.Sprintf("| %s = %s | ", name, value)
			}
		} else {
			values = "-"
		}

		// Todo: Need to update. Try to find and use some template to create and format msg
		formattedMsg := fmt.Sprintf(
			"\n" +
				"%s *%s*(%s) -> *%s*(%s) - %s\n" +
				"*%s*\n" +
				"*%s* *%s* *%s*\n" +
				"*%s*\n" +
				"Price: *%s* %s\n" +
				"Values: *%s*\n" +
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
			productPrice,
			price,
			// The new row
			values,
			// The new row
			newPost.RequestId,
			newPost.Dateup,
			newPost.DetailsPageUrl,
		)

		msg := tgbotapi.NewMessage(chatId, formattedMsg)
		msg.ParseMode = "markdown"

		// Prepare command
		command := core.BotInlineCommand{Method: "__create_post", RequestId: newPost.RequestId}
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

func alarmClock(bot *tgbotapi.BotAPI, chatId int64)  {
	intervalCh := time.Tick(time.Duration(core.Config.PingTimeout) * time.Second)
	for _ = range intervalCh {
		if core.IsWorking != true {
			return
		}
		msg := tgbotapi.NewMessage(chatId, "*God! How am I tired...*")
		msg.ParseMode = "markdown"
		_, _ = bot.Send(msg)
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

func stripValueName(content string) string {
	plainTex := content
	stripTagsReg := regexp.MustCompile(`([=:()]|\n|\s)*`)
	plainTex = stripTagsReg.ReplaceAllString(plainTex,"")
	plainTex = strings.Join(strings.Fields(plainTex), " ")
	return plainTex
}
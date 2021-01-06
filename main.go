package main

import (
	"encoding/json"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"log"
	"main/core"
	"main/handlers/inlineHandlers"
	"main/handlers/simpleHandlers"
	"net/http"
	"os"
)

func init()  {
	// Init first status of application
	core.IsWorking = false

	// Register all inline keyboard handlers
	core.RegisteredInlineCommands = map[string]interface{}{
		"__create_post": inlineHandlers.CreatePostInlineHandler,
	}

	// Register all simple chat handlers (SCommand - simple command)
	core.RegisteredSimpleCommands = map[string]interface{}{
		"stop": simpleHandlers.FinshSimpleHandler,
		"exit": simpleHandlers.ExitSimpleHandler,
		"help": simpleHandlers.HelpSimpleHandler,
		"sayhi": simpleHandlers.SayhiSimpleHandler,
		"status": simpleHandlers.StatusSimpleHandler,
		"start": simpleHandlers.StartSimpleHandler,
	}
}

func main() {
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
			var commandInline core.BotInlineCommand
			err := json.Unmarshal([]byte(update.CallbackQuery.Data), &commandInline)
			if err != nil {
				log.Println(err)
			}
			commandInline.RunCommand()
		}

		// Ignore any non-Message Updates
		if update.Message == nil {
			continue
		}

		// Simple command handler
		if update.Message.IsCommand() {
			// Create and set by default text for command answer
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.Text = "If you're so stupid, it's better to ask someone smarter. For example me /help"

			// Prepare simple command and execute
			commandSimple := core.BotSimpleCommand{update.Message.Command(), bot, update, &msg}
			commandSimple.RunCommand()

			log.Printf("The %s command was executed successful", update.Message.Command())
			_, _ = bot.Send(msg)
		}
	}
}

// Start page for bot on production Cloud
func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	_, _ = resp.Write([]byte("Hi all! I'm Telegram CargoBot on Heroku"))
}

// Todo: this code need to add when I want to use pager
//if href, ok := r.HTMLDoc.Find("li.next > a").Attr("href"); ok {
//	g.Get(r.JoinURL(href), quotesParse)
//}

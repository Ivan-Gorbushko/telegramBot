package simpleHandlers

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"main/core"
)

func FinshSimpleHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, msg *tgbotapi.MessageConfig)  {
	if core.IsWorking {
		core.IsWorking = false
		core.DisconnectMongo()
		msg.Text = "Fuh! Is it finally over..."
	} else {
		msg.Text = "You are silly I already don't work. And don't even think about running me!"
	}
}

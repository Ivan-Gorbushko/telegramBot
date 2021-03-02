package simpleHandlers

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"main/core"
	"os"
)

func ExitSimpleHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	core.DisconnectMongo()
	msg.Text = "Noooo you killed me!!!"
	_, _ = bot.Send(msg)
	// Exit successfully
	os.Exit(0)
}

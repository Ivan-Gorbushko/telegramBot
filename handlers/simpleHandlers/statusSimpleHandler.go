package simpleHandlers

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"main/core"
)

func StatusSimpleHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, msg *tgbotapi.MessageConfig)  {
	if core.IsWorking {
		msg.Text = "I'm working and I'm so busy to answer you"
	} else {
		msg.Text = "I do nothing but this is not a reason to work"
	}
}

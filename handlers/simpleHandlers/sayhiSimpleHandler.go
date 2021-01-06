package simpleHandlers

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

func SayhiSimpleHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, msg *tgbotapi.MessageConfig)  {
	msg.Text = "Hi bro:)"
}

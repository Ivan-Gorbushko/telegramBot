package simpleHandlers

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

func HelpSimpleHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, msg *tgbotapi.MessageConfig)  {
	msg.Text = "Ofc you can type:\n 1) /sayhi\n 2) /status\n 3) /start\n 4) /stop\n 5) /exit\n\nBut better leave me alone!"
}

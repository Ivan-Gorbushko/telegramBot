package core

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

var RegisteredInlineCommands map[string]interface{}
var RegisteredSimpleCommands map[string]interface{}

type BotInlineCommand struct {
	Method string
	RequestId string
}

type BotSimpleCommand struct {
	Method string
	Bot *tgbotapi.BotAPI
	Update tgbotapi.Update
	Msg *tgbotapi.MessageConfig
}

func (command BotInlineCommand) RunCommand() interface{} {
	if commandCallback, exist := RegisteredInlineCommands[command.Method]; exist {
		return commandCallback.(func(string) interface{})(command.RequestId)
	}

	return nil
}

func (command BotSimpleCommand) RunCommand() {
	if commandCallback, exist := RegisteredSimpleCommands[command.Method]; exist {
		commandCallback.(func(*tgbotapi.BotAPI, tgbotapi.Update, *tgbotapi.MessageConfig))(command.Bot, command.Update, command.Msg)
	}
}

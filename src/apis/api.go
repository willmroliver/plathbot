package apis

import (
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type API interface {
	Next(*botapi.BotAPI, any, string)
	SendMenu(*botapi.BotAPI, *botapi.Message)
}

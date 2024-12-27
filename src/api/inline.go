package api

import (
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type InlineAction func(*Context, *botapi.InlineQuery)

type InlineAPI struct {
	Actions map[string]InlineAction
}

func (api *InlineAPI) Select(c *Context, msg *botapi.InlineQuery, cmd string) {
	if action, exists := api.Actions[cmd]; exists {
		action(c, msg)
	}
}

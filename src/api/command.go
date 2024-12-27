package api

import (
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandAction func(*Context, *botapi.Message)

type CommandAPI struct {
	Actions map[string]CommandAction
}

func (api *CommandAPI) Select(c *Context, msg *botapi.Message, cmd string) {
	if action, exists := api.Actions[cmd]; exists {
		action(c, msg)
	}
}

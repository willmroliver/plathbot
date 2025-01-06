package api

import (
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandAction func(*Context, *botapi.Message, ...string)

type CommandAPI struct {
	Actions map[string]CommandAction
}

func (api *CommandAPI) Select(c *Context, msg *botapi.Message, args ...string) {
	if action, exists := api.Actions[args[0]]; exists {
		action(c, msg, args[1:]...)
	}
}

package api

import (
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandAction func(*Context, *botapi.Message, ...string)

type CommandAPI struct {
	Actions map[string]CommandAction
}

func (api *CommandAPI) Select(c *Context, msg *botapi.Message, args ...string) {
	action, ok := api.Actions[args[0]]
	if !ok && len(args[0]) > 2 && args[0][1] == 'p' {
		action, ok = api.Actions["/"+args[0][2:]]
	}

	if ok {
		action(c, msg, args[1:]...)
	}
}

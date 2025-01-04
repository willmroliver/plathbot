package api

import (
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandAction func(*Context, *botapi.Message)

type CommandAPI struct {
	Actions map[string]CommandAction
}

func (api *CommandAPI) Select(c *Context, msg *botapi.Message, cmd string) {
	i := strings.Index(cmd, " ")
	if i == -1 {
		i = len(cmd)
	}

	if action, exists := api.Actions[cmd[:i]]; exists {
		action(c, msg)
	}
}

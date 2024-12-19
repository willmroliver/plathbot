package apis

import (
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackCmd struct {
	cmd  string
	from int
}

func NewCallbackCmd(cmd string) *CallbackCmd {
	from := 0
	return &CallbackCmd{cmd, from}
}

func (cc *CallbackCmd) Path() string {
	return cc.cmd
}

func (cc *CallbackCmd) Get() string {
	to := strings.Index(cc.Tail(), "/")
	if to == -1 {
		to = len(cc.cmd)
	} else {
		to += cc.from
	}

	return cc.cmd[cc.from:to]
}

func (cc *CallbackCmd) Tail() string {
	return cc.cmd[cc.from:]
}

func (cc *CallbackCmd) Next() *CallbackCmd {
	index := strings.Index(cc.Tail(), "/")
	if index == -1 {
		return cc
	}

	cc.from = cc.from + index + 1
	return cc
}

type Callback map[string]func(*botapi.BotAPI, *botapi.CallbackQuery, *CallbackCmd)

func (api Callback) Next(bot *botapi.BotAPI, msg any, cmd *CallbackCmd) {
	if action, exists := api[cmd.Get()]; exists {
		action(bot, msg.(*botapi.CallbackQuery), cmd.Next())
	}
}

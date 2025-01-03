package api

import (
	"fmt"
	"log"
	"strings"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/util"
)

type CallbackAction func(*Context, *botapi.CallbackQuery, *CallbackCmd)

type CallbackConfig struct {
	Actions        map[string]CallbackAction
	DynamicActions func(*Context, *botapi.CallbackQuery, *CallbackCmd) map[string]CallbackAction
	DynamicOptions func(*Context, *botapi.CallbackQuery, *CallbackCmd) []map[string]string
	PublicOptions  []map[string]string
	PublicCooldown time.Duration
	PublicOnly     bool
	PrivateOptions []map[string]string
	PrivateOnly    bool
}

type CallbackAPI struct {
	Title          string
	Actions        map[string]CallbackAction
	DynamicActions func(*Context, *botapi.CallbackQuery, *CallbackCmd) map[string]CallbackAction
	DynamicOptions func(*Context, *botapi.CallbackQuery, *CallbackCmd) []map[string]string
	PublicOptions  []map[string]string
	PublicCooldown time.Duration
	PublicOnly     bool
	PrivateOptions []map[string]string
	PrivateOnly    bool
	Path           string
}

func NewCallbackAPI(title, path string, config *CallbackConfig) (api *CallbackAPI) {
	api = &CallbackAPI{
		Title:          title,
		Actions:        config.Actions,
		DynamicActions: config.DynamicActions,
		DynamicOptions: config.DynamicOptions,
		PublicOptions:  config.PublicOptions,
		PublicCooldown: config.PublicCooldown,
		PublicOnly:     config.PublicOnly,
		PrivateOptions: config.PrivateOptions,
		PrivateOnly:    config.PrivateOnly,
		Path:           path,
	}

	if config.DynamicOptions == nil {
		api.PublicOptions = api.resolveOpts(api.PublicOptions)
		api.PrivateOptions = api.resolveOpts(api.PrivateOptions)
	}

	return
}

func (api *CallbackAPI) Select(c *Context, q *botapi.CallbackQuery, cc *CallbackCmd) {
	if api.DynamicActions != nil {
		api.Actions = api.DynamicActions(c, q, cc)
	}

	if api.DynamicOptions != nil {
		api.PublicOptions = api.resolveOpts(api.DynamicOptions(c, q, cc))
	}

	if action, exists := api.Actions[cc.Get()]; exists {
		action(c, q, cc.Next())
		return
	}

	api.Expose(c, q)
}

func (api *CallbackAPI) Expose(c *Context, m *botapi.CallbackQuery) {
	private := c.Chat.Type == "private"

	if !private && !util.TryLockFor(fmt.Sprintf("%d %s", c.Chat.ID, api.Title), api.PublicCooldown) {
		return
	}

	if api.PrivateOnly && !private {
		msg := botapi.NewMessage(
			c.Chat.ID,
			fmt.Sprintf("DM to use %q - %s", api.Title, util.AtBotString(c.Bot)),
		)
		msg.ParseMode = "Markdown"
		util.SendConfig(c.Bot, &msg)

		return
	}

	opts := &api.PublicOptions
	if private && !api.PublicOnly && api.DynamicOptions == nil {
		opts = &api.PrivateOptions
	}

	var err error

	if m == nil {
		msg := botapi.NewMessage(c.Chat.ID, api.Title)
		msg.ReplyMarkup = *util.InlineKeyboard(*opts)

		_, err = c.Bot.Send(msg)
	} else {
		msg := botapi.NewEditMessageText(c.Chat.ID, m.Message.MessageID, api.Title)
		msg.ReplyMarkup = util.InlineKeyboard(*opts)

		err = util.SendUpdate(c.Bot, &msg)
	}

	if err != nil {
		log.Printf("Error sending %q menu: %q", api.Title, err.Error())
	}
}

func (api *CallbackAPI) resolveOpts(opts []map[string]string) (res []map[string]string) {
	res = opts

	for i, row := range opts {
		for k, v := range row {
			if strings.HasPrefix(v, "_") {
				continue
			}

			if v == ".." {
				if j := strings.LastIndex(api.Path, "/"); j != -1 {
					res[i][k] = strings.Join([]string{api.Path[:j], v}, "/")
				} else {
					res[i][k] = "/"
				}

				continue
			}

			res[i][k] = strings.Join([]string{api.Path, v}, "/")
		}
	}

	return res
}

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
		cc.from = len(cc.cmd)
		return cc
	}

	cc.from = cc.from + index + 1
	return cc
}

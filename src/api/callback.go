package api

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	BuiltInDelete = "_DEL"
)

var (
	builtin = map[string]func(*Context, *botapi.CallbackQuery){
		BuiltInDelete: func(ctx *Context, q *botapi.CallbackQuery) {
			u := botapi.NewDeleteMessage(ctx.Chat.ID, q.Message.MessageID)
			SendConfig(ctx.Bot, &u)
		},
	}
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
	Path           string
	Actions        map[string]CallbackAction
	DynamicActions func(*Context, *botapi.CallbackQuery, *CallbackCmd) map[string]CallbackAction
	DynamicOptions func(*Context, *botapi.CallbackQuery, *CallbackCmd) []map[string]string
	PublicOptions  []map[string]string
	PublicCooldown time.Duration
	PublicOnly     bool
	PrivateOptions []map[string]string
	PrivateOnly    bool
}

func NewCallbackAPI(title, path string, config *CallbackConfig) (api *CallbackAPI) {
	api = &CallbackAPI{
		Title:          title,
		Path:           path,
		Actions:        config.Actions,
		DynamicActions: config.DynamicActions,
		DynamicOptions: config.DynamicOptions,
		PublicOptions:  config.PublicOptions,
		PublicCooldown: config.PublicCooldown,
		PublicOnly:     config.PublicOnly,
		PrivateOptions: config.PrivateOptions,
		PrivateOnly:    config.PrivateOnly,
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

	if action, exists := api.Actions[cc.Get()]; exists {
		action(c, q, cc.Next())
		return
	}

	if s, ok := cc.Tags["user"]; ok {
		if userID, _ := strconv.ParseInt(s, 10, 64); userID != c.User.ID {
			return
		}
	}

	if cb, ok := builtin[cc.Path()]; ok {
		go cb(c, q)
		return
	}

	api.Expose(c, q, cc)
}

func (api *CallbackAPI) Expose(c *Context, q *botapi.CallbackQuery, cc *CallbackCmd) {
	private := c.Chat.Type == "private"

	if !private && !util.TryLockFor(fmt.Sprintf("%d %s", c.Chat.ID, api.Title), api.PublicCooldown) {
		return
	}

	if api.PrivateOnly && !private {
		msg := &botapi.MessageConfig{
			BaseChat: botapi.BaseChat{ChatID: c.Chat.ID},
		}

		msg.ReplyMarkup = *InlineKeyboard([]map[string]string{{
			fmt.Sprintf(
				"DM to use %q - %s",
				api.Title,
				AtBotString(c.Bot),
			): KeyboardLink(ToPrivateString(c.Bot, api.Path)),
		}})

		msg.ParseMode = "Markdown"
		SendConfig(c.Bot, msg)

		return
	}

	opts := &api.PublicOptions

	if api.DynamicOptions != nil {
		api.PublicOptions = api.resolveOpts(api.DynamicOptions(c, q, cc))
	} else if private && !api.PublicOnly {
		opts = &api.PrivateOptions
	}

	var err error

	if q == nil {
		msg := botapi.NewMessage(c.Chat.ID, api.Title)
		msg.ReplyMarkup = *InlineKeyboard(*opts, fmt.Sprintf("user=%d", c.User.ID))

		_, err = c.Bot.Send(msg)
	} else {
		msg := botapi.NewEditMessageText(c.Chat.ID, q.Message.MessageID, api.Title)
		msg.ReplyMarkup = InlineKeyboard(*opts, fmt.Sprintf("user=%d", c.User.ID))

		err = SendUpdate(c.Bot, &msg)
	}

	if err != nil {
		log.Printf("Error sending %q menu: %q", api.Title, err.Error())
	}
}

func (api *CallbackAPI) resolveOpts(opts []map[string]string) (res []map[string]string) {
	res = opts

	for i, row := range opts {
		for k, v := range row {
			if strings.HasPrefix(v, "_") || strings.HasPrefix(v, "!!") {
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

			if api.Path == "" {
				res[i][k] = v
			} else {
				res[i][k] = strings.Join([]string{api.Path, v}, "/")
			}
		}
	}

	return res
}

type CallbackCmd struct {
	cmd  string
	from int
	Tags map[string]string
}

// NewCallbackCmd creates a new CallbackCmd, parsing the 'cmd' string for a prefix of
// comma-separated tags. These can be individual values or key=value pairs.
//
//   - "action/to/something/" 					- No tags
//   - "tag1,tag2 action/to/something/" 		- Value tags
//   - "user=158,chat=20 action/to/something/" 	- Key-value tags
//   - "user=158,tag2 action/to/something/" 	- Mixed tags
func NewCallbackCmd(cmd string) *CallbackCmd {
	tags := map[string]string{}
	from := 0

	i := strings.Index(cmd, "|")
	if i != -1 && i < len(cmd)-1 {
		vals := cmd[:i]
		cmd = cmd[i+1:]

		for _, v := range strings.Split(vals, ",") {
			if i = strings.Index(v, "="); i != -1 {
				tags[v[:i]] = v[i+1:]
			} else {
				tags[v] = v
			}
		}
	}

	return &CallbackCmd{cmd, from, tags}
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

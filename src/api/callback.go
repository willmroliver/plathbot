package api

import (
	"fmt"
	"log"
	"maps"
	"strconv"
	"strings"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	BuiltInDelete = "_DEL"
)

var builtin = map[string]func(*Context, *botapi.CallbackQuery){
	BuiltInDelete: func(ctx *Context, q *botapi.CallbackQuery) {
		u := botapi.NewDeleteMessage(ctx.Chat.ID, q.Message.MessageID)
		SendConfig(ctx.Bot, &u)
	},
}

type CallbackAction func(*Context, *botapi.CallbackQuery, *CallbackCmd)

type CallbackExtension struct {
	title, cmd string
	action     CallbackAction
}

type CallbackExtensions []*CallbackExtension

func (exts *CallbackExtensions) ExtendAPI(title, cmd string, action CallbackAction) {
	*exts = append(*exts, &CallbackExtension{title, cmd, action})
}

func (exts *CallbackExtensions) Options() []map[string]string {
	opts := make([]map[string]string, len(*exts))

	for i, ext := range *exts {
		opts[i] = map[string]string{ext.title: ext.cmd}
	}

	return opts
}

type CallbackConfig struct {
	Actions        map[string]CallbackAction
	DynamicActions func(*Context, *botapi.CallbackQuery, *CallbackCmd) map[string]CallbackAction
	DynamicOptions func(*Context, *botapi.CallbackQuery, *CallbackCmd) []map[string]string
	PublicOptions  []map[string]string
	PublicCooldown time.Duration
	PublicOnly     bool
	PrivateOptions []map[string]string
	PrivateOnly    bool
	Extensions     CallbackExtensions
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
	Extensions     CallbackExtensions
}

func NewCallbackAPI(title, path string, config *CallbackConfig) (api *CallbackAPI) {
	if config.Actions == nil {
		config.Actions = make(map[string]CallbackAction)
	}

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
		Extensions:     config.Extensions,
	}

	for _, ext := range api.Extensions {
		api.Actions[ext.cmd] = ext.action

		if !api.PublicOnly {
			api.PrivateOptions = append(api.PrivateOptions, map[string]string{ext.title: ext.cmd})
		}
		if !api.PrivateOnly {
			api.PublicOptions = append(api.PublicOptions, map[string]string{ext.title: ext.cmd})
		}
	}

	if config.DynamicOptions == nil {
		api.PublicOptions = api.resolveOpts(api.PublicOptions)
		api.PrivateOptions = api.resolveOpts(api.PrivateOptions)
	}

	return
}

func (api *CallbackAPI) Select(c *Context, q *botapi.CallbackQuery, cc *CallbackCmd) {
	if api.PrivateOnly && c.Chat.Type != "private" {
		api.privateRedirect(c, q)
		return
	}

	if s, ok := cc.Tags["user"]; ok {
		if userID, _ := strconv.ParseInt(s, 10, 64); userID != c.User.ID {
			log.Printf("User %d cannot use keyboard owned by %d\n", c.User.ID, userID)
			return
		}
	}

	if api.DynamicActions != nil {
		maps.Copy(api.Actions, api.DynamicActions(c, q, cc))
	}

	var cmd string

	if cmd = cc.Get(); cmd == "" {
		api.Expose(c, q, cc)
		return
	}

	if action, ok := api.Actions[cmd]; ok {
		log.Printf("Direct match in %s: %s\n", api.Title, cmd)
		action(c, q, cc.Next())
		return
	}

	if _, ok := cc.Tags["cmd"]; ok {
		test := strings.ToLower(cmd)

		for key, action := range api.Actions {
			if strings.Contains(strings.ToLower(key), test) {
				log.Printf("Partial match in %s: %s -> %s\n", api.Title, cmd, key)
				action(c, q, cc.Next())
				return
			}
		}
	}

	if cmd == "help" || cmd == "?" {
		api.SendHelp(c, q, cc)
		return
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

	opts := &api.PublicOptions

	if api.DynamicOptions != nil {
		api.PublicOptions = api.resolveOpts(
			append(
				api.Extensions.Options(),
				api.DynamicOptions(c, q, cc)...,
			),
		)
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

func (api *CallbackAPI) SendHelp(c *Context, q *botapi.CallbackQuery, cc *CallbackCmd) {
	path := "/" + strings.ReplaceAll(api.Path, "/", " ")
	root := path == "/"

	text := &strings.Builder{}
	text.WriteString("*" + api.Title + "* - Commands:\n\n")

	if !root {
		path += " "
	}

	for key := range api.Actions {
		if strings.HasPrefix(key, "_") {
			continue
		}

		text.WriteString(fmt.Sprintf("\t\t\t\t%s%s\n", path, key))
	}

	if root {
		text.WriteString(`
You can access most sub-menus using just commands.
			*/stats games week*

To see available sub-commands, use:
			*/cmd help*, or 
			*/cmd ?*
		`)
	} else if api.DynamicOptions != nil {
		text.WriteString(fmt.Sprintf(`
Partial matches are supported!

So, if the command text is 
				"🚀 Space stuff"
You could use:
				%sspace				%s🚀
		`, path, path))
	}

	if q == nil {
		msg := &botapi.MessageConfig{
			BaseChat: botapi.BaseChat{ChatID: c.Chat.ID},
		}
		msg.ParseMode = botapi.ModeMarkdown
		msg.Text = text.String()
		SendConfig(c.Bot, msg)
	} else {
		msg := botapi.NewEditMessageText(c.Chat.ID, q.Message.MessageID, api.Title)
		msg.ParseMode = botapi.ModeMarkdown
		msg.Text = text.String()
		SendUpdate(c.Bot, &msg)
	}
}

func (api *CallbackAPI) privateRedirect(c *Context, q *botapi.CallbackQuery) {
	dest := api.Path
	if i := strings.Index(api.Path, "/"); i != -1 {
		dest = dest[:i]
	}

	text := "🤫 Shhh.. You're in public"
	mu := *InlineKeyboard([]map[string]string{{
		api.Title: KeyboardLink(ToPrivateString(c.Bot, dest)),
	}})

	if q == nil {
		msg := &botapi.MessageConfig{
			BaseChat: botapi.BaseChat{ChatID: c.Chat.ID},
		}
		msg.ParseMode = "Markdown"
		msg.Text = text
		msg.ReplyMarkup = mu
		SendConfig(c.Bot, msg)
	} else {
		msg := botapi.NewEditMessageTextAndMarkup(c.Chat.ID, q.Message.MessageID, api.Title, mu)
		msg.ParseMode = "Markdown"
		msg.Text = text
		SendUpdate(c.Bot, &msg)
	}
}

func (api *CallbackAPI) resolveOpts(opts []map[string]string) (res []map[string]string) {
	res = make([]map[string]string, 0, len(opts))

	i := -1

	for _, row := range opts {
		if len(row) == 0 {
			continue
		}

		i++
		res = append(res, row)

		for k, v := range res[i] {
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
//   - action/to/something/ 					- No tags
//   - tag1,tag2|action/to/something/ 			- Value tags
//   - user=158,chat=20|action/to/something/ 	- Key-value tags
//   - user=158,tag2|action/to/something/ 		- Mixed tags
func NewCallbackCmd(cmd string) *CallbackCmd {
	tags := map[string]string{}
	from := 0

	i := strings.Index(cmd, "|")
	if i != -1 && i < len(cmd)-1 {
		vals := cmd[:i]
		cmd = cmd[i+1:]

		for v := range strings.SplitSeq(vals, ",") {
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

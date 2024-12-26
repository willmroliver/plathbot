package apis

import (
	"fmt"
	"log"
	"strings"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/server"
	"github.com/willmroliver/plathbot/src/util"
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

type CallbackAction func(*server.Server, *botapi.CallbackQuery, *CallbackCmd)

type Callback struct {
	Title          string
	Actions        map[string]CallbackAction
	PublicOptions  []map[string]string
	PublicCooldown time.Duration
	PublicOnly     bool
	PrivateOptions []map[string]string
	PrivateOnly    bool
}

func (api *Callback) Select(s *server.Server, msg *botapi.CallbackQuery, cmd *CallbackCmd) {
	if action, exists := api.Actions[cmd.Get()]; exists {
		action(s, msg, cmd.Next())
		return
	}

	api.Expose(s, msg.Message.Chat, msg.Message)
}

func (api *Callback) Expose(s *server.Server, chat *botapi.Chat, msg *botapi.Message) {
	if chat == nil && msg == nil {
		return
	}

	if chat == nil {
		chat = msg.Chat
	}

	send := func(c *botapi.Chat, m *botapi.Message, opts []map[string]string) {
		var err error

		if m == nil {
			msg := botapi.NewMessage(c.ID, api.Title)
			msg.ReplyMarkup = util.InlineKeyboard(opts)

			_, err = s.Bot.Send(msg)
		} else {
			options := util.InlineKeyboard(opts)
			msg := botapi.NewEditMessageText(m.Chat.ID, m.MessageID, api.Title)
			msg.ReplyMarkup = &options

			err = util.SendUpdate(s.Bot, &msg)
		}

		if err != nil {
			log.Printf("Error sending %q menu: %q", api.Title, err.Error())
		}
	}

	private := chat.Type == "private"

	if !private && !util.TryLockFor(fmt.Sprintf("%d %s", chat.ID, api.Title), api.PublicCooldown) {
		return
	}

	if !private || api.PublicOnly {
		if api.PrivateOnly {
			msg := botapi.NewMessage(
				chat.ID,
				fmt.Sprintf("DM to use %q - %s", api.Title, util.AtBotString(s.Bot)),
			)
			msg.ParseMode = "Markdown"
			util.SendConfig(s.Bot, &msg)

			return
		}

		send(chat, nil, api.PublicOptions)
		return
	}

	send(chat, msg, api.PrivateOptions)
}

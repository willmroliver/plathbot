package server

import (
	"log"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/core"
)

type Context struct {
	bot    *botapi.BotAPI
	update botapi.Update
}

func NewContext(bot *botapi.BotAPI, update botapi.Update) *Context {
	return &Context{
		bot,
		update,
	}
}

func (ctx *Context) HandleUpdate() {
	ctx.HandleMessage()
	ctx.HandleMessageReaction()
	ctx.HandleCallbackQuery()
	ctx.HandleInlineQuery()
}

func (ctx *Context) HandleMessage() {
	m := ctx.update.Message
	if m == nil {
		return
	}

	from := m.From
	text := m.Text

	if from == nil {
		return
	}

	log.Printf("Received %q from %s", text, from.FirstName)

	if m.Chat.Type != "private" {
		cmd, valid := strings.CutPrefix(text, "/plath@")
		if !valid {
			return
		}

		text = "/" + cmd
	}

	var err error

	switch {
	case strings.HasPrefix(text, "/"):
		err = core.HandleCommand(ctx.bot, m, text)
	default:
		break
	}

	if err != nil {
		log.Printf("An error ocurred: %s", err.Error())
	}
}

func (ctx *Context) HandleMessageReaction() {
	m := ctx.update.MessageReaction
	if m == nil {
		return
	}

	u := m.User
	if u == nil {
		return
	}

	reacts := map[string][]*botapi.ReactionType{
		"was": m.OldReaction,
		"is":  m.NewReaction,
	}

	for label, values := range reacts {
		value := ""
		if len(values) > 0 {
			value = values[0].Emoji
		}

		log.Printf("%s: %s", strings.ToTitle(label), value)
	}
}

func (ctx *Context) HandleCallbackQuery() {
	m := ctx.update.CallbackQuery
	if m == nil {
		return
	}

	core.HandleCallbackQuery(ctx.bot, ctx.update.CallbackQuery, m.Data)
}

func (ctx *Context) HandleInlineQuery() {
	if m := ctx.update.InlineQuery; m != nil {
		if err := core.HandleInlineQuery(ctx.bot, m); err != nil {
			log.Printf("HandleInlineQuery error: %q", err.Error())
		}
	}
}

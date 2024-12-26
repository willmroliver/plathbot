package core

import (
	"errors"
	"log"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/server"
	"gorm.io/gorm"
)

type Context struct {
	server *server.Server
	update *botapi.Update
}

func NewContext(server *server.Server, update *botapi.Update) *Context {
	return &Context{
		server,
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

	if m.Chat.Type != "private" {
		ctx.addXP(from, 10)

		cmd, valid := strings.CutPrefix(text, "/plath@")
		if !valid {
			return
		}

		text = "/" + cmd
	}

	var err error

	switch {
	case strings.HasPrefix(text, "/"):
		err = handleCommand(ctx.server, m, text)
	default:
		break
	}

	if err != nil {
		log.Printf("An error ocurred: %s", err.Error())
	}
}

func (ctx *Context) HandleMessageReaction() {
	m := ctx.update.MessageReaction
	if m == nil || m.User == nil || m.Chat.Type == "private" {
		return
	}

	switch {
	case len(m.OldReaction) < len(m.NewReaction):
		ctx.addXP(m.User, 10)
	case len(m.OldReaction) > len(m.NewReaction):
		ctx.addXP(m.User, -10)
	default:
		break
	}
}

func (ctx *Context) HandleCallbackQuery() {
	m := ctx.update.CallbackQuery
	if m == nil {
		return
	}

	handleCallbackQuery(ctx.server, ctx.update.CallbackQuery, m.Data)
}

func (ctx *Context) HandleInlineQuery() {
	if m := ctx.update.InlineQuery; m != nil {
		if err := handleInlineQuery(ctx.server, m); err != nil {
			log.Printf("HandleInlineQuery error: %q", err.Error())
		}
	}
}

func (ctx *Context) addXP(u *botapi.User, xp int) (user *model.User, err error) {
	user = model.NewUser(u)

	err = ctx.server.DB.
		Model(u).
		Where("id = ?", u.ID).
		UpdateColumn("xp", gorm.Expr("xp + ?", xp)).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error opening user %d record: %q", u.ID, err.Error())
	}

	return
}

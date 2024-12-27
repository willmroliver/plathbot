package api

import (
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
)

type Context struct {
	Server   *Server
	Bot      *botapi.BotAPI
	Update   *botapi.Update
	UserRepo *repo.UserRepo
	User     *botapi.User
	Chat     *botapi.Chat
	Message  *botapi.Message
}

func NewContext(server *Server, update *botapi.Update) *Context {
	return &Context{
		Server:   server,
		Bot:      server.Bot,
		Update:   update,
		UserRepo: repo.NewUserRepo(server.DB),
	}
}

func (ctx *Context) HandleUpdate() {
	ctx.HandleMessage()
	ctx.HandleMessageReaction()
	ctx.HandleCallbackQuery()
	ctx.HandleInlineQuery()
}

func (ctx *Context) HandleMessage() {
	m := ctx.Update.Message
	if m == nil || m.From == nil {
		return
	}

	ctx.User = m.From
	ctx.Chat = m.Chat
	ctx.Message = m

	text := m.Text

	if m.Chat.Type != "private" {
		ctx.UserRepo.ShiftXP(ctx.User, 10)
	}

	switch {
	case strings.HasPrefix(text, "/"):
		ctx.Server.CommandAPI.Select(ctx, m, text)
	default:
		break
	}
}

func (ctx *Context) HandleMessageReaction() {
	m := ctx.Update.MessageReaction
	if m == nil || m.User == nil || m.Chat.Type == "private" {
		return
	}

	ctx.User = m.User
	ctx.Chat = m.Chat
	ctx.Message = m

	switch {
	case len(m.OldReaction) < len(m.NewReaction):
		ctx.UserRepo.ShiftXP(ctx.User, 10)
	case len(m.OldReaction) > len(m.NewReaction):
		ctx.UserRepo.ShiftXP(ctx.User, -10)
	default:
		break
	}
}

func (ctx *Context) HandleCallbackQuery() {
	m := ctx.Update.CallbackQuery
	if m == nil || m.From == nil {
		return
	}

	ctx.User = m.From
	ctx.Chat = m.Message.Chat
	ctx.Message = m.Message

	go ctx.Server.CallbackAPI.Select(ctx, m, NewCallbackCmd(m.Data))
}

func (ctx *Context) HandleInlineQuery() {
	if m := ctx.Update.InlineQuery; m != nil {
		ctx.Server.InlineAPI.Select(ctx, m, m.Query)
	}
}

func (ctx *Context) GetUser() *model.User {
	return ctx.UserRepo.Get(ctx.User)
}

package api

import (
	"fmt"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/service"
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
		service.
			NewUserXPService(ctx.Server.DB).
			UpdateXPs(ctx.User, service.XPTitleEngage, 10)
	}

	if !strings.HasPrefix(text, "/") {
		return
	}

	text = strings.Replace(text, "@"+ctx.Bot.Self.UserName, "", 1)

	var cc *CallbackCmd

	if strings.HasPrefix(text, "/start ") {
		cc = NewCallbackCmd(fmt.Sprintf("user=%d,cmd|", ctx.User.ID) + strings.ReplaceAll(text[len("/start "):], " ", "/") + "/")
	} else {
		cc = NewCallbackCmd(fmt.Sprintf("user=%d,cmd|", ctx.User.ID) + strings.ReplaceAll(text, " ", "/")[1:] + "/")
	}

	cmd := cc.Get()

	action, ok := ctx.Server.CallbackAPI.Actions[cmd]
	if !ok && cmd[0] == 'p' {
		action, ok = ctx.Server.CallbackAPI.Actions[cmd[1:]]
	}

	if ok {
		msg, err := SendBasic(ctx.Bot, ctx.Chat.ID, "🚀")

		if err == nil {
			ctx.Message = msg
			q := &botapi.CallbackQuery{
				From:    m.From,
				Message: msg,
			}

			action(ctx, q, cc.Next())
			return
		}
	}

	ctx.Server.CommandAPI.Select(ctx, m, strings.Split(text, " ")...)
}

func (ctx *Context) HandleMessageReaction() {
	m := ctx.Update.MessageReaction
	if m == nil || m.User == nil || m.Chat.Type == "private" {
		return
	}

	ctx.User = m.User
	ctx.Chat = m.Chat
	ctx.Message = m

	reactService := service.NewReactService(ctx.Server.DB)
	reactService.UpdateCounts(m)

	switch {
	case len(m.OldReaction) < len(m.NewReaction):
		service.
			NewUserXPService(ctx.Server.DB).
			UpdateXPs(ctx.User, service.XPTitleEngage, 10)
	case len(m.OldReaction) > len(m.NewReaction):
		service.
			NewUserXPService(ctx.Server.DB).
			UpdateXPs(ctx.User, service.XPTitleEngage, -10)
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

	ctx.Server.CallbackAPI.Select(ctx, m, NewCallbackCmd(m.Data))
}

func (ctx *Context) HandleInlineQuery() {
	if m := ctx.Update.InlineQuery; m != nil {
		ctx.Server.InlineAPI.Select(ctx, m, m.Query)
	}
}

func (ctx *Context) GetUser() *model.User {
	return ctx.UserRepo.Get(ctx.User)
}

func (ctx *Context) IsAdmin() bool {
	if ctx.Chat == nil {
		return false
	}

	return ctx.Chat.Type == "private" ||
		ctx.UserRepo.Get(ctx.User).IsAdmin(ctx.Bot, ctx.Chat.ID)
}

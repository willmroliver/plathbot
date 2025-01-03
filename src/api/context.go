package api

import (
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/service"
	"github.com/willmroliver/plathbot/src/util"
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

	if i := strings.LastIndex(text, "@"); i != -1 && text[i+1:] != ctx.Bot.Self.UserName {
		return
	} else if i != -1 {
		text = text[:i]
	}

	ctx.Server.CommandAPI.Select(ctx, m, text)
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

	basic := map[string]func(*Context, *botapi.CallbackQuery){
		"_delete": func(ctx *Context, q *botapi.CallbackQuery) {
			u := botapi.NewDeleteMessage(ctx.Chat.ID, q.Message.MessageID)
			util.SendConfig(ctx.Bot, &u)
		},
	}

	if cb, ok := basic[m.Data]; ok {
		go cb(ctx, m)
		return
	}

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

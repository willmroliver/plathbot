package api

import (
	"errors"
	"log"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"gorm.io/gorm"
)

type Context struct {
	Server   *Server
	Bot      *botapi.BotAPI
	Update   *botapi.Update
	UserRepo *repo.UserRepo
	User     *model.User
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
	if m == nil {
		return
	}

	if err := ctx.trySetUser(m.From); err != nil {
		return
	}

	ctx.Chat = m.Chat
	ctx.Message = m

	text := m.Text

	if m.Chat.Type != "private" {
		ctx.addXP(ctx.User.TelegramUser, 10)

		cmd, valid := strings.CutPrefix(text, "/plath@")
		if !valid {
			return
		}

		text = "/" + cmd
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
	m := ctx.Update.CallbackQuery
	if m == nil {
		return
	}

	if err := ctx.trySetUser(m.From); err != nil {
		return
	}

	ctx.Chat = m.Message.Chat
	ctx.Message = m.Message

	go ctx.Server.CallbackAPI.Select(ctx, m, NewCallbackCmd(m.Data))
}

func (ctx *Context) HandleInlineQuery() {
	if m := ctx.Update.InlineQuery; m != nil {
		ctx.Server.InlineAPI.Select(ctx, m, m.Query)
	}
}

func (ctx *Context) trySetUser(user *botapi.User) (err error) {
	if user == nil {
		err = errors.New("trySetUser() - No user passed")
		return
	}

	ctx.User = model.NewUser(user)
	err = ctx.UserRepo.Get(ctx.User, ctx.User.ID)
	return
}

func (ctx *Context) addXP(u *botapi.User, xp int) (user *model.User, err error) {
	user = model.NewUser(u)

	err = ctx.Server.DB.
		Model(u).
		Where("id = ?", u.ID).
		UpdateColumn("xp", gorm.Expr("xp + ?", xp)).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error opening user %d record: %q", u.ID, err.Error())
	}

	return
}

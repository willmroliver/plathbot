package emoji

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/service"
	"github.com/willmroliver/plathbot/src/util"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

const (
	AdminTitle = "🔐 Manage"
	AdminPath  = Path + "/admin"
)

var open = sync.Map{}

func AdminAPI() *api.CallbackAPI {
	return api.NewCallbackAPI(
		AdminTitle,
		AdminPath,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"view": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if a := OpenAdmin(c.Server.DB, cq); a != nil {
						a.View(c, cq)
					}
				},
				"update": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if a := OpenAdmin(c.Server.DB, cq); a != nil {
						a.Update(c, cq)
					}
				},
				"remove": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if a := OpenAdmin(c.Server.DB, cq); a != nil {
						a.Remove(c, cq)
					}
				},
			},
			PublicOptions: []map[string]string{
				{"✏️ Update": "update"},
				{"👀 View": "view", "🗑️ Remove": "remove"},
				util.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

type Admin struct {
	*api.Interaction[string]
	service *service.ReactService
	user    *botapi.User
}

func NewAdmin(db *gorm.DB, q *botapi.CallbackQuery) *Admin {
	return &Admin{
		api.NewInteraction(q.Message, ""),
		service.NewReactService(db),
		q.From,
	}
}

func OpenAdmin(db *gorm.DB, q *botapi.CallbackQuery) (admin *Admin) {
	open.Range(func(key any, value any) bool {
		if value.(*api.Interaction[any]).Age() > time.Minute*5 {
			open.Delete(key)
		}

		return true
	})

	if data, exists := open.Load(q.From.ID); exists {
		admin = data.(*Admin)
	} else {
		admin = NewAdmin(db, q)
	}

	return
}

func (a *Admin) User() *model.User {
	return a.service.UserRepo.Get(a.user)
}

func (a *Admin) View(c *api.Context, query *botapi.CallbackQuery) {
	reactRepo := repo.NewReactRepo(c.Server.DB)
	tracked := reactRepo.All()

	text := &strings.Builder{}
	text.WriteString("Currently tracked:\n\n")

	for _, react := range tracked {
		text.WriteString(fmt.Sprintf("%s - %s\n", react.Emoji, react.Title))
	}

	util.SendUpdate(c.Bot, a.NewMessageUpdate(text.String(), &[]map[string]string{
		util.KeyboardNavRow(AdminPath),
	}))
}

func (a *Admin) Update(c *api.Context, query *botapi.CallbackQuery) {
	util.SendBasic(c.Bot, c.Chat.ID, `
Okay, send the emoji you'd like to update and give it a title, space-separated.
E.g: '💸 High-flyer'`)

	hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, data any) {
		a := data.(*Admin)
		if m.From.ID != a.user.ID {
			return
		}

		i := strings.Index(m.Text, " ")
		if i == -1 || !util.IsEmoji(m.Text[:i]) || i+1 == len(m.Text) {
			return
		}

		e, t := m.Text[:i], m.Text[i+1:]
		if a.service.ReactRepo.Save(e, t) != nil {
			return
		}

		util.SendConfig(s.Bot, a.NewMessage(fmt.Sprintf("%s saved as %q", e, t), &[]map[string]string{
			util.KeyboardNavRow(AdminPath),
		}))
	}, a, time.Minute*5)

	c.Server.RegisterMessageHook(c.Chat.ID, hook)
}

func (a *Admin) Remove(c *api.Context, query *botapi.CallbackQuery) {
	util.SendBasic(c.Bot, c.Chat.ID, `
Okay, send the emoji you'd like to stop tracking.`)

	hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, userID any) {
		if m.From.ID != userID.(int64) {
			return
		}

		e := m.Text

		if i := strings.Index(m.Text, " "); i > 0 {
			if e = m.Text[:i]; !util.IsEmoji(e) {
				return
			}
		}

		if reactService := service.NewReactService(s.DB); reactService.Untrack(e) != nil {
			return
		}

		util.SendConfig(s.Bot, a.NewMessage(fmt.Sprintf("%s removed", e), &[]map[string]string{
			util.KeyboardNavRow(AdminPath),
		}))
	}, c.User.ID, time.Minute*5)

	c.Server.RegisterMessageHook(c.Chat.ID, hook)
}

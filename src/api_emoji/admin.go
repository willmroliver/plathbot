package emoji

import (
	"fmt"
	"strconv"
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
					if a := OpenAdmin(c, cq, cc); a != nil {
						a.View(c, cq)
					}
				},
				"update": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if a := OpenAdmin(c, cq, cc); a != nil {
						a.Update(c, cq)
					}
				},
				"remove": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if a := OpenAdmin(c, cq, cc); a != nil {
						a.Remove(c, cq)
					}
				},
			},
			PublicOptions: []map[string]string{
				{"✏️ Update": "update"},
				{"👀 View": "view", "🗑️ Remove": "remove"},
				api.KeyboardNavRow(".."),
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

func OpenAdmin(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) (admin *Admin) {
	if u, ok := cc.Tags["user"]; ok {
		if id, _ := strconv.ParseInt(u, 10, 64); id != q.From.ID {
			return
		}
	}

	if !c.IsAdmin() {
		return
	}

	open.Range(func(key any, value any) bool {
		if value.(*api.Interaction[any]).Age() > time.Minute*5 {
			open.Delete(key)
		}

		return true
	})

	if data, exists := open.Load(q.From.ID); exists {
		admin = data.(*Admin)
	} else {
		admin = NewAdmin(c.Server.DB, q)
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
		text.WriteString(react.Emoji + " - " + react.Title + "\n")
	}

	api.SendUpdate(c.Bot, a.NewMessageUpdate(text.String(), api.InlineKeyboard([]map[string]string{
		api.KeyboardNavRow(AdminPath),
	}, fmt.Sprintf("user=%d", a.user.ID))))
}

func (a *Admin) Update(c *api.Context, query *botapi.CallbackQuery) {
	api.SendBasic(c.Bot, c.Chat.ID, `
Okay, send the emoji you'd like to update and give it a title, space-separated.
E.g: '💸 High-flyer'`)

	hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, data any) (done bool) {
		done = true

		ad := data.(*Admin)
		if m.From.ID != ad.user.ID {
			return
		}

		i := strings.Index(m.Text, " ")
		if i == -1 || !util.IsEmoji(m.Text[:i]) || i+1 == len(m.Text) {
			return
		}

		e, t := util.NormalizeEmoji(m.Text[:i]), m.Text[i+1:]
		if ad.service.ReactRepo.Save(e, t) != nil {
			return
		}

		mu := api.InlineKeyboard([]map[string]string{
			api.KeyboardNavRow(AdminPath),
		}, fmt.Sprintf("user=%d", ad.user.ID))

		api.SendConfig(s.Bot, ad.NewMessage(fmt.Sprintf("%s saved as %q", e, t), mu))

		return
	}, a, time.Minute*5)

	c.Server.RegisterMessageHook(c.Chat.ID, hook)
}

func (a *Admin) Remove(c *api.Context, query *botapi.CallbackQuery) {
	api.SendBasic(c.Bot, c.Chat.ID, `
Okay, send the emoji you'd like to stop tracking.`)

	hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, userID any) (done bool) {
		if m.From.ID != userID.(int64) {
			return
		}

		done = true

		e := m.Text

		if i := strings.Index(m.Text, " "); i > 0 {
			if e = m.Text[:i]; !util.IsEmoji(e) {
				return
			}
		}

		if reactService := service.NewReactService(s.DB); reactService.Untrack(e) != nil {
			return
		}

		mu := api.InlineKeyboard([]map[string]string{
			api.KeyboardNavRow(AdminPath),
		}, fmt.Sprintf("user=%d", userID))

		api.SendConfig(s.Bot, a.NewMessage(e+" removed", mu))

		return
	}, c.User.ID, time.Minute*5)

	c.Server.RegisterMessageHook(c.Chat.ID, hook)
}

package emoji

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/willmroliver/plathbot/src/api"
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
					reactRepo := repo.NewReactRepo(c.Server.DB)
					tracked := reactRepo.All()

					text := &strings.Builder{}
					text.WriteString("Currently tracked:\n\n")

					for _, react := range tracked {
						text.WriteString(fmt.Sprintf("%s - %s\n", react.Emoji, react.Title))
					}

					i := api.NewInteraction[bool](cq.Message, true)

					util.SendUpdate(c.Bot, i.NewMessageUpdate(text.String(), &[]map[string]string{
						{"👈 Back": AdminPath},
					}))
				},
				"update": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					util.SendBasic(c.Bot, c.Chat.ID, `
Okay, send the emoji you'd like to update and give it a title, space-separated.
E.g: '💸 High-flyer'`)

					hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, userID any) {
						if m.From.ID != userID.(int64) {
							return
						}

						i := strings.Index(m.Text, " ")
						if i == -1 || !util.IsEmoji(m.Text[:i]) || i+1 == len(m.Text) {
							return
						}

						e, t := m.Text[:i], m.Text[i+1:]

						reactRepo := repo.NewReactRepo(c.Server.DB)

						if reactRepo.Save(e, m.Text[i+1:]) != nil {
							return
						}

						in := api.NewInteraction[bool](cq.Message, true)

						util.SendConfig(s.Bot, in.NewMessage(fmt.Sprintf("%s saved as %q", e, t), &[]map[string]string{
							{"👈 Back": AdminPath},
						}))
					}, c.User.ID, time.Minute*5)

					c.Server.RegisterMessageHook(c.Chat.ID, hook)
				},
				"remove": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
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

						i := api.NewInteraction[bool](cq.Message, true)

						util.SendConfig(s.Bot, i.NewMessage(fmt.Sprintf("%s removed", e), &[]map[string]string{
							{"👈 Back": AdminPath},
						}))
					}, c.User.ID, time.Minute*5)

					c.Server.RegisterMessageHook(c.Chat.ID, hook)
				},
			},
			PublicOptions: []map[string]string{
				{"✏️ Update": "update"},
				{"👀 View": "view", "🗑️ Remove": "remove"},
				{"👈 Back": ".."},
			},
			PublicOnly: true,
		},
	)
}

type Admin struct {
	*api.Interaction[string]
	service *service.ReactService
}

func NewAdmin(db *gorm.DB, q *botapi.CallbackQuery) *Admin {
	return &Admin{
		api.NewInteraction(q.Message, ""),
		service.NewReactService(db),
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

func (a *Admin) View(c *api.Context, query *botapi.CallbackQuery) {
	reactRepo := repo.NewReactRepo(c.Server.DB)
	tracked := reactRepo.All()

	text := &strings.Builder{}
	text.WriteString("Currently tracked:\n\n")

	for _, react := range tracked {
		text.WriteString(fmt.Sprintf("%s - %s\n", react.Emoji, react.Title))
	}

	util.SendUpdate(c.Bot, a.NewMessageUpdate(text.String(), &[]map[string]string{
		{"👈 Back": AdminPath},
	}))
}

func (a *Admin) Update(c *api.Context, query *botapi.CallbackQuery) {

}

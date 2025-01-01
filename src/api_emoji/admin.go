package emoji

import (
	"fmt"
	"strings"
	"time"

	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/service"
	"github.com/willmroliver/plathbot/src/util"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	AdminTitle = "🔐 Manage"
	AdminPath  = Path + "/admin"
)

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
					m := util.InlineKeyboard([]map[string]string{{"👈 Back": AdminPath}})
					u := i.NewMessageUpdate(text.String(), &m)

					util.SendUpdate(c.Bot, &u)
				},
				"update": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					util.SendBasic(c.Bot, c.Chat.ID, `
Okay, send the emoji you'd like to update and give it a title, space-separated.
E.g: '💸 High-flyer'`)

					hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, userID any) bool {
						if m.From.ID != userID.(int64) {
							return false
						}

						i := strings.Index(m.Text, " ")
						if i == -1 || !util.IsEmoji(m.Text[:i]) || i+1 == len(m.Text) {
							return false
						}

						e, t := m.Text[:i], m.Text[i+1:]

						reactRepo := repo.NewReactRepo(c.Server.DB)

						if reactRepo.Save(e, m.Text[i+1:]) == nil {
							i := api.NewInteraction[bool](cq.Message, true)
							m := util.InlineKeyboard([]map[string]string{{"👈 Back": AdminPath}})
							u := i.NewMessage(fmt.Sprintf("%s saved as %q", e, t), &m)

							util.SendConfig(s.Bot, &u)
							return true
						}

						return false
					}, c.User.ID, time.Minute*5)

					c.Server.RegisterMessageHook(c.Chat.ID, hook)
				},
				"remove": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					util.SendBasic(c.Bot, c.Chat.ID, `
Okay, send the emoji you'd like to stop tracking.`)

					hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, userID any) bool {
						if m.From.ID != userID.(int64) {
							return false
						}

						e := m.Text

						if i := strings.Index(m.Text, " "); i > 0 {
							if e = m.Text[:i]; !util.IsEmoji(e) {
								return false
							}
						}

						reactService := service.NewReactService(s.DB)

						if reactService.Untrack(e) == nil {
							util.SendBasic(s.Bot, m.Chat.ID, fmt.Sprintf("%s removed", e))
							return true
						}

						return false
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

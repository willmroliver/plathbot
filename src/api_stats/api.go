//go:build stats
// +build stats

package stats

import (
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/repo"
)

const (
	Title = "📊 Stats"
	Path  = "stats"
)

var (
	Extensions api.CallbackExtensions
)

func init() {
	api.RegisterCallbackAPI(Path, API)
}

func API() *api.CallbackAPI {
	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			DynamicActions: func(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) (actions map[string]api.CallbackAction) {
				actions = make(map[string]api.CallbackAction)

				if titles := repo.NewUserXPRepo(c.Server.DB).Titles(); titles != nil {
					for _, title := range titles {
						actions[title] = UserXPAPI(title).Select
					}
				}

				return
			},
			DynamicOptions: func(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) (options []map[string]string) {
				if titles := repo.NewUserXPRepo(c.Server.DB).Titles(); titles != nil {
					options = make([]map[string]string, len(titles)+2)

					for i, title := range titles {
						options[i] = map[string]string{title: title}
					}

					options[len(options)-1] = api.KeyboardNavRow("..")

					return
				}

				return []map[string]string{api.KeyboardNavRow("..")}
			},
		},
	)
}

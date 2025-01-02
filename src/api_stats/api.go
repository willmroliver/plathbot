package stats

import (
	"time"

	"github.com/willmroliver/plathbot/src/api"
	emoji "github.com/willmroliver/plathbot/src/api_emoji"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	Title = "📊 Stats"
	Path  = "stats"
)

var (
	emojiAPI = emoji.TableAPI()
	xpAPI    = UserXPAPI()
)

func API() *api.CallbackAPI {
	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"xp":    xpAPI.Select,
				"emoji": emojiAPI.Select,
			},
			PublicCooldown: time.Second * 3,
			PublicOptions: []map[string]string{
				{XpTitle: "xp"},
				{emoji.Title: "emoji"},
				util.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

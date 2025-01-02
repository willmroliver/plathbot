package stats

import (
	"time"

	"github.com/willmroliver/plathbot/src/api"
	emoji "github.com/willmroliver/plathbot/src/api_emoji"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	Title = "🙂 Emojis"
	Path  = "emojis"
)

var (
	emojiAPI = emoji.TableAPI()
	xpAPI    = XpAPI()
)

func API() *api.CallbackAPI {
	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"emoji": emojiAPI.Select,
				"xp":    xpAPI.Select,
			},
			PublicCooldown: time.Second * 3,
			PublicOptions: []map[string]string{
				{emoji.TableTitle: "table"},
				{XpTitle: "admin"},
				util.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

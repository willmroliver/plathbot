package games

import (
	"time"

	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	Title = "🎮 Games"
	Path  = "games"
)

func API() *api.CallbackAPI {
	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"cointoss": CointossQuery,
			},
			PublicCooldown: time.Second * 3,
			PublicOptions: []map[string]string{
				{CointossTitle: "cointoss"},
				util.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

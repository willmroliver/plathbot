package games

import (
	"time"

	"github.com/willmroliver/plathbot/src/api"
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
			PublicCooldown: time.Second * 15,
			PublicOptions: []map[string]string{
				{"🪙 Cointoss": "cointoss"},
			},
			PublicOnly: true,
		},
	)
}

package games

import (
	"time"

	"github.com/willmroliver/plathbot/src/apis"
)

const (
	Title = "🎮 Games"
	Path  = "games"
)

func API() *apis.Callback {
	return apis.NewCallback(
		Title,
		Path,
		&apis.CallbackConfig{
			Actions: map[string]apis.CallbackAction{
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

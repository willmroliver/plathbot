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
				"cointoss":          CointossQuery,
				"rockpaperscissors": RockPaperScissorsQuery,
			},
			PublicCooldown: time.Second * 3,
			PublicOptions: []map[string]string{
				{CointossTitle: "cointoss"},
				{RockPaperScissorsTitle: "rockpaperscissors"},
				api.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

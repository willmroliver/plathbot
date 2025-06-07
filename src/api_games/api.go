//go:build games
// +build games

package games

import (
	"time"

	"github.com/willmroliver/plathbot/src/api"
)

const (
	Title = "🎮 Games"
	Path  = "games"
)

func init() {
	api.RegisterCallbackAPI(Path, API)
}

func API() *api.CallbackAPI {
	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"cointoss":          CointossQuery,
				"rockpaperscissors": RockPaperScissorsQuery,
				"connect4":          ConnectFourQuery,
			},
			PublicCooldown: time.Second * 3,
			PublicOptions: []map[string]string{
				{CointossTitle: "cointoss"},
				{RockPaperScissorsTitle: "rockpaperscissors"},
				{ConnectFourTitle: "connect4"},
				api.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

package emoji

import (
	"time"

	"github.com/willmroliver/plathbot/src/api"
)

const (
	Title = "🙂 Emojis"
	Path  = "emojis"
)

var (
	adminAPI = AdminAPI()
)

func API() *api.CallbackAPI {
	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"admin": adminAPI.Select,
			},
			PublicCooldown: time.Second * 3,
			PublicOptions: []map[string]string{
				{AdminTitle: "admin"},
				{"👈 Back": ".."},
			},
			PublicOnly: true,
		},
	)
}

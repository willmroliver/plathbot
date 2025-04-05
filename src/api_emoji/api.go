package emoji

import (
	"time"

	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	Title = "🙂 Emojis"
	Path  = "emojis"
)

var (
	tableAPI = TableAPI()
	adminAPI = AdminAPI()
)

func API() *api.CallbackAPI {
	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"table": tableAPI.Select,
				"admin": adminAPI.Select,
			},
			PublicCooldown: time.Second * 3,
			PublicOptions: []map[string]string{
				{TableTitle: "table"},
				{AdminTitle: "admin"},
				util.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

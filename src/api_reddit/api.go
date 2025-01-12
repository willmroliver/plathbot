package reddit

import (
	"time"

	"github.com/willmroliver/plathbot/src/api"
)

const (
	Title = "🤖 Reddit"
	Path  = "reddit"
)

var (
	adminAPI = AdminAPI()
)

func API() *api.CallbackAPI {
	admin, _ := "admin", "view"

	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				admin: adminAPI.Select,
			},
			PublicCooldown: time.Second * 3,
			PublicOptions: []map[string]string{
				{AdminTitle: admin},
				api.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

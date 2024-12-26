package games

import (
	"fmt"
	"time"

	"github.com/willmroliver/plathbot/src/apis"
)

func API() *apis.Callback {
	return &apis.Callback{
		Title: "🎮 Games",
		Actions: map[string]apis.CallbackAction{
			"cointoss": CointossQuery,
		},
		PublicCooldown: time.Second * 15,
		PublicOptions: []map[string]string{
			{"🪙 Cointoss": getCmd("cointoss")},
		},
		PublicOnly: true,
	}
}

func getCmd(name string) string {
	return fmt.Sprintf("games/%s/", name)
}

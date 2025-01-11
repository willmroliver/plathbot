package account

import (
	"time"

	"github.com/willmroliver/plathbot/src/api"
)

const (
	Title = "💻 Your Account"
	Path  = "account"
)

var (
	walletAPI = WalletAPI()
	redditAPI = RedditAPI()
)

func API() *api.CallbackAPI {
	wallet, reddit, xp := "wallet", "reddit", "xp"

	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				wallet: walletAPI.Select,
				reddit: redditAPI.Select,
				xp:     XPQuery,
			},
			PublicCooldown: time.Second * 15,
			PrivateOptions: []map[string]string{
				{WalletTitle: wallet},
				{RedditTitle: reddit},
				{XPTitle: xp},
				api.KeyboardNavRow(".."),
			},
			PrivateOnly: true,
		},
	)
}

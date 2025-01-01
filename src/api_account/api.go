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
)

func API() *api.CallbackAPI {
	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"wallet": walletAPI.Select,
				"xp":     XPQuery,
			},
			PublicCooldown: time.Second * 15,
			PrivateOptions: []map[string]string{
				{WalletTitle: "wallet"},
				{XPTitle: "xp"},
				{"👈 Back": ".."},
			},
			PrivateOnly: true,
		},
	)
}

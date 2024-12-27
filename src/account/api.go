package account

import (
	"time"

	"github.com/willmroliver/plathbot/src/apis"
)

const (
	Title = "💻 Your Account"
	Path  = "account"
)

var (
	walletAPI = WalletAPI()
)

func API() *apis.Callback {
	return apis.NewCallback(
		Title,
		Path,
		&apis.CallbackConfig{
			Actions: map[string]apis.CallbackAction{
				"wallet": walletAPI.Select,
				"xp":     XPQuery,
			},
			PublicCooldown: time.Second * 15,
			PrivateOptions: []map[string]string{
				{WalletTitle: "wallet"},
				{"📈 My XP": "xp"},
			},
			PrivateOnly: true,
		},
	)
}

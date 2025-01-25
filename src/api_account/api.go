//go:build account
// +build account

package account

import (
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
)

const (
	Title = "💻 Account"
	Path  = "account"
)

var (
	walletAPI = WalletAPI()

	Extensions api.CallbackExtensions
)

func init() {
	api.RegisterCallbackAPI(Path, API)
}

func API() *api.CallbackAPI {
	wallet := "wallet"

	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			DynamicActions: func(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) (opts map[string]api.CallbackAction) {
				opts = map[string]api.CallbackAction{
					wallet: walletAPI.Select,
				}

				return
			},
			DynamicOptions: func(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) (opts []map[string]string) {
				opts = []map[string]string{
					{WalletTitle: wallet},
					api.KeyboardNavRow(".."),
				}

				return
			},
			PrivateOnly: true,
			Extensions:  Extensions,
		},
	)
}

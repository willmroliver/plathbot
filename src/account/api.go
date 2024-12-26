package account

import (
	"fmt"
	"time"

	"github.com/willmroliver/plathbot/src/apis"
)

func API() *apis.Callback {
	return &apis.Callback{
		Title: "💻 Your Account",
		Actions: map[string]apis.CallbackAction{
			"wallet": WalletQuery,
			"xp":     XPQuery,
		},
		PublicCooldown: time.Second * 15,
		PrivateOptions: []map[string]string{
			{"💳 Wallet": getCmd("wallet")},
			{"📈 My XP": getCmd("xp")},
		},
		PrivateOnly: true,
	}
}

func getCmd(name string) string {
	return fmt.Sprintf("account/%s/", name)
}

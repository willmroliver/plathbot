package account

import (
	"log"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/apis"
)

func WalletQuery(bot *botapi.BotAPI, query *botapi.CallbackQuery, cmd *apis.CallbackCmd) {
	log.Println("Edit wallet request")
}

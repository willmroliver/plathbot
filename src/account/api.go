package account

import (
	"fmt"
	"log"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/apis"
	"github.com/willmroliver/plathbot/src/util"
)

func HandleCallbackQuery(bot *botapi.BotAPI, m *botapi.CallbackQuery, cmd *apis.CallbackCmd) {
	api := apis.Callback{
		"wallet": WalletQuery,
	}

	api.Next(bot, m, cmd)
}

func SendOptions(bot *botapi.BotAPI, m *botapi.Message) (err error) {
	if m.Chat.Type != "private" {
		if !util.TryLockFor(fmt.Sprintf("%d games", m.Chat.ID), time.Second*15) {
			return nil
		}

		msg := botapi.NewMessage(m.Chat.ID, "DM to manage your account - "+util.AtBotString(bot))
		msg.ParseMode = "Markdown"

		util.SendConfig(bot, &msg)
	}

	msg := botapi.NewMessage(m.Chat.ID, "Your Account")
	msg.ReplyMarkup = util.InlineKeyboard([]map[string]string{
		{"Wallet": getCmd("wallet")},
	})

	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("Error sending menu: %q", err.Error())
	}

	return
}

func getCmd(name string) string {
	return fmt.Sprintf("account/%s/", name)
}

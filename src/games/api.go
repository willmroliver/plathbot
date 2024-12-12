package games

import (
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var api = map[string]func(*botapi.BotAPI, *botapi.CallbackQuery, string){
	"cointoss": CointossQuery,
}

func HandleCallbackQuery(bot *botapi.BotAPI, m *botapi.CallbackQuery, cmd string) {
	for key, action := range api {
		if strings.HasPrefix(cmd, key+"/") {
			action(bot, m, cmd[len(key)+1:])
			break
		}
	}
}

func SendMenu(bot *botapi.BotAPI, chatID int64) (err error) {
	msg := botapi.NewMessage(chatID, "Want to play?")
	msg.ReplyMarkup = botapi.NewInlineKeyboardMarkup(
		botapi.NewInlineKeyboardRow(
			botapi.NewInlineKeyboardButtonData("Coin Toss", "games/cointoss/"),
		),
	)

	_, err = bot.Send(msg)

	return
}

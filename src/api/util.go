package api

import (
	"fmt"
	"log"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	keyboardLinkCode   = "!!l"
	keyboardSwitchCode = "!!s"
)

func SendBasic(bot *botapi.BotAPI, chatID int64, msg string) (err error) {
	_, err = bot.Send(botapi.NewMessage(chatID, msg))
	if err != nil {
		log.Printf("Got error: %q, attempting to send %q:", err.Error(), msg)
	}

	return
}

func SendConfig(bot *botapi.BotAPI, msg botapi.Chattable) (err error) {
	if _, err = bot.Send(msg); err != nil {
		log.Printf("Got error: %q, attempting to send %+v:", err.Error(), msg)
	}

	return
}

func SendUpdate(bot *botapi.BotAPI, msg *botapi.EditMessageTextConfig) (err error) {
	if _, err = bot.Send(*msg); err != nil {
		log.Printf("Got error: %q, attempting to send %q", err.Error(), msg.Text)
	}

	return
}

func RequestBasic(bot *botapi.BotAPI, query *botapi.InlineQuery, title, msg string) (err error) {
	a := botapi.NewInlineQueryResultArticleHTML(query.ID, title, msg)
	c := botapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     3600,
		Results:       []interface{}{a},
	}

	_, err = bot.Request(c)
	return
}

func InlineKeyboard(data []map[string]string) *botapi.InlineKeyboardMarkup {
	rows := make([][]botapi.InlineKeyboardButton, len(data))

	for i, row := range data {
		buttons := make([]botapi.InlineKeyboardButton, len(row))
		j := 0

		for text, data := range row {
			buttons[j] = KeyboardButton(text, data)
			j++
		}

		rows[i] = botapi.NewInlineKeyboardRow(buttons...)
	}
	res := botapi.NewInlineKeyboardMarkup(rows...)
	return &res
}

// KeyboardSwitch creates a string key with embedded information for InlineKeyboard() to generate an inline query switch button
func KeyboardInlineSwitch(query string) string {
	return fmt.Sprintf("%s(%s)", keyboardSwitchCode, query)
}

func KeyboardLink(url string) string {
	return fmt.Sprintf("%s(%s)", keyboardLinkCode, url)
}

// KeyboardButton interprets a button data string to support in-built functions that select particular button types.
//
// Data must be of the form `!!x(arg1,arg2,...)`, where 'x' is some character selecting a button-type code, and the args
// are passed to the `NewInlineKeyboardButton[Data/Switch/...]()` func
//
// Supported currently:
//
// !!s - NewInlineKeyboardButtonSwitch(text, query)
// !!l - NewInlineKeyboardButtonURL(text, url)
func KeyboardButton(text, data string) botapi.InlineKeyboardButton {
	ops := map[string]func(text string, args ...string) botapi.InlineKeyboardButton{
		keyboardSwitchCode: func(text string, args ...string) botapi.InlineKeyboardButton {
			return botapi.NewInlineKeyboardButtonSwitch(text, args[0])
		},
		keyboardLinkCode: func(text string, args ...string) botapi.InlineKeyboardButton {
			return botapi.NewInlineKeyboardButtonURL(text, args[0])
		},
	}

	if strings.HasPrefix(data, "!!") {
		if i := strings.Index(data, "("); i != -1 {
			if op := ops[data[:i]]; op != nil {
				return op(text, strings.Split(data[i+1:len(data)-1], ",")...)
			}
		}
	}

	return botapi.NewInlineKeyboardButtonData(text, data)
}

func KeyboardNavRow(back string) map[string]string {
	return map[string]string{"👈 Back": back, "👋 Done": BuiltInDelete}
}

func AtString(text string, id int64) string {
	return fmt.Sprintf("[%s](tg://user?id=%d)", text, id)
}

func AtUserString(user *botapi.User) string {
	if user == nil {
		return ""
	}

	return fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, user.ID)
}

func AtBotString(bot *botapi.BotAPI) string {
	return fmt.Sprintf("[@%s](https://t.me/%s)", bot.Self.UserName, bot.Self.UserName)
}

func ToPrivateString(bot *botapi.BotAPI, cmd string) string {
	return fmt.Sprintf("tg://resolve?domain=%s&start=%s", bot.Self.UserName, cmd)
}

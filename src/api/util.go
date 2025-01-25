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

func SendBasic(bot *botapi.BotAPI, chatID int64, msg string) (*botapi.Message, error) {
	m, err := bot.Send(botapi.NewMessage(chatID, msg))
	if err != nil {
		log.Printf("SendBasic error: %q, attempting to send %q:", err.Error(), msg)
	}

	return &m, err
}

func SendConfig(bot *botapi.BotAPI, msg botapi.Chattable) (*botapi.Message, error) {
	m, err := bot.Send(msg)
	if err != nil {
		log.Printf("SendConfig error: %q, attempting to send %+v:", err.Error(), msg)
	}

	return &m, err
}

func SendUpdate(bot *botapi.BotAPI, msg *botapi.EditMessageTextConfig) (err error) {
	if _, err = bot.Send(*msg); err != nil {
		log.Printf("SendUpdate error: %q, attempting to send %q", err.Error(), msg.Text)
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

// InlineKeyboard converts an array of Text:Data maps into a TG inline keyboard.
// Data supports functions which can request special button types.
//
// Optional tags can be passed which are prepended as a comma-separated list: "data" -> "arg1,arg2,... data"
func InlineKeyboard(data []map[string]string, tags ...string) *botapi.InlineKeyboardMarkup {
	rows := make([][]botapi.InlineKeyboardButton, len(data))

	for i, row := range data {
		buttons := make([]botapi.InlineKeyboardButton, len(row))
		j := 0

		for text, data := range row {
			buttons[j] = KeyboardButton(text, data, tags...)
			j++
		}

		rows[i] = botapi.NewInlineKeyboardRow(buttons...)
	}

	res := botapi.NewInlineKeyboardMarkup(rows...)
	return &res
}

// KeyboardSwitch creates a string key with embedded information
// for KeyboardButton(result) to generate an inline query switch button
func KeyboardInlineSwitch(query string) string {
	return keyboardSwitchCode + "(" + query + ")"
}

// KeyboardSwitch creates a string key with embedded information
// for KeyboardButton(result) to generate a URL button
func KeyboardLink(url string) string {
	return keyboardLinkCode + "(" + url + ")"
}

// KeyboardButton interprets a button data string to support in-built functions
// that select particular button types, or to prefix tags to a data-button payload.
//
// Data must be of the form `!!x(arg1,arg2,...)`,
// where 'x' is some character selecting a button-type code, and the args
// are passed to the `NewInlineKeyboardButton[Data/Switch/...]()` func
//
// Supported currently:
//   - !!s(text, query) - NewInlineKeyboardButtonSwitch(text, url)\n
//   - !!l(text, url) - NewInlineKeyboardButtonURL(text, url)
func KeyboardButton(text, data string, tags ...string) botapi.InlineKeyboardButton {
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

	if prefix := strings.Join(tags, ","); prefix != "" {
		data = prefix + "|" + data
	}

	return botapi.NewInlineKeyboardButtonData(text, data)
}

func KeyboardNavRow(back string) map[string]string {
	return map[string]string{"👈 Back": back, "👋 Done": BuiltInDelete}
}

func AtString(text string, id int64) string {
	return fmt.Sprintf("[%s](tg://user?id=%d)", text, id)
}

func DisplayName(user *botapi.User) (text string) {
	if user.FirstName != "" {
		text = user.FirstName
	} else if user.UserName != "" {
		text = user.UserName
	} else {
		text = fmt.Sprintf("%d", user.ID)
	}

	return
}

func AtUserString(user *botapi.User) string {
	if user == nil {
		return ""
	}

	var text string

	if user.FirstName != "" {
		text = user.FirstName
	} else if user.UserName != "" {
		text = user.UserName
	} else {
		text = fmt.Sprintf("user-%d", user.ID)
	}

	return fmt.Sprintf("[%s](tg://user?id=%d)", text, user.ID)
}

func AtBotString(bot *botapi.BotAPI) string {
	return "[@" + bot.Self.UserName + "](https://t.me/" + bot.Self.UserName + ")"
}

func ToPrivateString(bot *botapi.BotAPI, cmd string) string {
	return fmt.Sprintf("tg://resolve?domain=%s&start=%s", bot.Self.UserName, cmd)
}

func MarkdownV2Cols(items []string, cols int) (res string) {
	maxes := make([]int, cols)

	for i := range len(items) {
		if n := len(items[i]); n > maxes[i%cols] {
			maxes[i%cols] = n
		}
	}

	for i := range maxes {
		maxes[i] += 2
	}

	res += "```\n"

	for i, item := range items {
		res += item + strings.Repeat(" ", maxes[i%cols]-len(item))

		if i%cols == cols-1 {
			res += "\n"
		}
	}

	res += "\n```"

	return
}

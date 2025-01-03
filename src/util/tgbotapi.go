package util

import (
	"fmt"
	"log"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
			buttons[j] = botapi.NewInlineKeyboardButtonData(text, data)
			j++
		}

		rows[i] = botapi.NewInlineKeyboardRow(buttons...)
	}
	res := botapi.NewInlineKeyboardMarkup(rows...)
	return &res
}

func KeyboardNavRow(back string) map[string]string {
	return map[string]string{"👈 Back": back, "👋 Done": "_delete"}
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

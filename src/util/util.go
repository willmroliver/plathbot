package util

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendBasic(bot *botapi.BotAPI, chatID int64, msg string) (err error) {
	_, err = bot.Send(botapi.NewMessage(chatID, msg))
	if err != nil {
		log.Printf("Got error: %q, attempting to send %q:", err.Error(), msg)
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

func AtUser(user *botapi.User) string {
	if user == nil {
		return ""
	}

	return fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, user.ID)
}

func PseudoRandInt(n int, seed bool) int {
	if seed {
		rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return rand.Int() % n
}

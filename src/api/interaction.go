package api

import (
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/util"
)

type Interaction[T comparable] struct {
	state T
	time  time.Time
	msg   *botapi.Message
}

func NewInteraction[T comparable](msg *botapi.Message, state T) *Interaction[T] {
	return &Interaction[T]{state, time.Now(), msg}
}

func (i *Interaction[T]) Mutate(state T, m *botapi.Message) {
	i.state = state
	i.time = time.Now()
	i.msg = m
}

func (i *Interaction[T]) Is(state T) bool {
	return i.state == state
}

func (i *Interaction[T]) Age() time.Duration {
	return time.Since(i.time)
}

func (i *Interaction[T]) NewMessage(text string, markup *[]map[string]string) *botapi.MessageConfig {
	msg := botapi.NewMessage(i.msg.Chat.ID, text)
	msg.ParseMode = botapi.ModeMarkdown

	if markup != nil {
		msg.ReplyMarkup = util.InlineKeyboard(*markup)
	}

	return &msg
}

func (i *Interaction[T]) NewMessageUpdate(text string, markup *[]map[string]string) *botapi.EditMessageTextConfig {
	var msg botapi.EditMessageTextConfig
	if markup != nil {
		msg = botapi.NewEditMessageTextAndMarkup(i.msg.Chat.ID, i.msg.MessageID, text, *util.InlineKeyboard(*markup))
	} else {
		msg = botapi.NewEditMessageText(i.msg.Chat.ID, i.msg.MessageID, text)
	}

	msg.ParseMode = botapi.ModeMarkdown
	return &msg
}

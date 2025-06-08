package api

import (
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Interaction[T comparable] struct {
	Msg *botapi.Message

	state T
	time  time.Time
}

func NewInteraction[T comparable](msg *botapi.Message, state T) *Interaction[T] {
	return &Interaction[T]{msg, state, time.Now()}
}

func (i *Interaction[T]) Mutate(state T, m *botapi.Message) {
	i.Msg = m

	i.state = state
	i.time = time.Now()
}

func (i *Interaction[T]) Is(state T) bool {
	return i.state == state
}

func (i *Interaction[T]) Age() time.Duration {
	return time.Since(i.time)
}

func (i *Interaction[T]) NewMessage(text string, markup *tgbotapi.InlineKeyboardMarkup) *botapi.MessageConfig {
	msg := botapi.NewMessage(i.Msg.Chat.ID, text)
	msg.ParseMode = botapi.ModeMarkdown

	if markup != nil {
		msg.ReplyMarkup = *markup
	}

	return &msg
}

func (i *Interaction[T]) NewMessageUpdate(text string, markup *tgbotapi.InlineKeyboardMarkup) *botapi.EditMessageTextConfig {
	var msg botapi.EditMessageTextConfig
	if markup != nil {
		msg = botapi.NewEditMessageTextAndMarkup(
			i.Msg.Chat.ID,
			i.Msg.MessageID,
			text,
			*markup,
		)
	} else {
		msg = botapi.NewEditMessageText(i.Msg.Chat.ID, i.Msg.MessageID, text)
	}

	msg.ParseMode = botapi.ModeMarkdown
	return &msg
}

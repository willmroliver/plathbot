package apis

import (
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Interaction[T comparable] struct {
	state     T
	time      time.Time
	chatID    int64
	messageID int
}

func NewInteraction[T comparable](msg *botapi.Message, state T) *Interaction[T] {
	return &Interaction[T]{state, time.Now(), msg.Chat.ID, msg.MessageID}
}

func (i *Interaction[T]) Mutate(state T, m *botapi.Message) {
	i.state = state
	i.time = time.Now()
	i.messageID = m.MessageID
}

func (i *Interaction[T]) Is(state T) bool {
	return i.state == state
}

func (i *Interaction[T]) Age() time.Duration {
	return time.Since(i.time)
}

func (i *Interaction[T]) NewMessage(text string) botapi.MessageConfig {
	msg := botapi.NewMessage(i.chatID, text)
	msg.ParseMode = botapi.ModeMarkdown
	return msg
}

func (i *Interaction[T]) NewMessageUpdate(text string, markup botapi.InlineKeyboardMarkup) botapi.EditMessageTextConfig {
	msg := botapi.NewEditMessageTextAndMarkup(i.chatID, i.messageID, text, markup)
	msg.ParseMode = botapi.ModeMarkdown
	return msg
}

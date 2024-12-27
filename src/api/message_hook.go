package api

import (
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHookable func(*Server, *botapi.Message, any)

type MessageHook struct {
	ExpiresAt time.Time
	Hook      MessageHookable
	Data      any
}

func NewMessageHook(cb MessageHookable, data any, lifespan time.Duration) *MessageHook {
	if lifespan > time.Hour {
		lifespan = time.Hour
	}

	return &MessageHook{
		ExpiresAt: time.Now().Add(lifespan),
		Hook:      cb,
		Data:      data,
	}
}

func (h *MessageHook) Execute(s *Server, m *botapi.Message) {
	h.Hook(s, m, h.Data)
}

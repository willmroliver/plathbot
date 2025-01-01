package api

import (
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageHookable is the accepted signature of message hook callbacks.
//
// Returning false indicates that the hook was not successful and should persist.
type MessageHookable func(*Server, *botapi.Message, any) bool

type MessageHook struct {
	ExpiresAt time.Time
	Hook      MessageHookable
	Data      any
}

// NewMessageHook packs a hookable callback with a payload to pass on execution, and a lifespan.
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

// Execute executes the hookable, passing the given *Server, *Message and Data payload.
func (h *MessageHook) Execute(s *Server, m *botapi.Message) bool {
	return h.Hook(s, m, h.Data)
}

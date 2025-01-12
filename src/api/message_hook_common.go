package api

import (
	"strings"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetDurationHook(chatID int64, lifespan time.Duration) (*MessageHook, chan time.Duration) {
	ch := make(chan time.Duration)

	return NewMessageHook(func(s *Server, m *botapi.Message, data any) (done bool) {
		dur, err := time.ParseDuration(strings.ReplaceAll(m.Text, " ", ""))
		if err != nil {
			SendBasic(s.Bot, chatID, "Invalid duration. Accepted units are 'h', 'm' and 's'.")
			return
		}

		ch <- dur

		done = true
		return
	}, nil, lifespan), ch
}

func GetTimeHook(chatID int64, lifespan time.Duration) (*MessageHook, chan time.Time) {
	errText := `<pre>
Could not get a valid time from your message. Format should be like:

- 20 Jul 99					(Assumes 00:00 +0000)	
- 20 Jul 99 07:00			(Assumes UTC)
- 20 Jul 99 07:00 BST		(British Summer Time)
- 20 Jul 99 07:00 +0100	    (UTC+1)
</pre>`
	errMsg := botapi.NewMessage(chatID, errText)
	errMsg.ParseMode = botapi.ModeHTML

	ch := make(chan time.Time)

	return NewMessageHook(func(s *Server, m *botapi.Message, data any) (done bool) {
		switch {
		case strings.Count(m.Text, " ") < 3:
			SendConfig(s.Bot, errMsg)
			return
		case strings.Count(m.Text, " ") == 3:
			m.Text += " 00:00 +0000"
		case strings.Count(m.Text, " ") == 4:
			m.Text += " +0000"
		default:
			break
		}

		t, err := time.Parse(time.RFC822, m.Text)
		if err != nil {
			SendConfig(s.Bot, errMsg)
			return
		}

		ch <- t

		done = true
		return
	}, nil, lifespan), ch
}

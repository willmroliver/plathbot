package api

import (
	"bufio"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/util"
	"gorm.io/gorm"
)

type Server struct {
	Bot         *botapi.BotAPI
	DB          *gorm.DB
	CallbackAPI *CallbackAPI
	CommandAPI  *CommandAPI
	InlineAPI   *InlineAPI

	chatHooks sync.Map
	userHooks sync.Map
}

func NewServer(db *gorm.DB) *Server {
	// Create a custom HTTP transport with DialContext
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Prefer IPv4
			return (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext(ctx, "tcp4", addr)
		},
		// Optional: additional transport settings
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// Create a custom HTTP client with the transport
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	bot, err := botapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Client = httpClient

	return &Server{
		Bot: bot,
		DB:  db,
	}
}

func (s *Server) Listen() {
	u := botapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{
		"message",
		"callback_query",
		"inline_query",
		"message_reaction",
		"message_reaction_count",
	}

	updates := s.Bot.GetUpdatesChan(u)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	listen := func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-updates:
				go func() {
					if s.DoMessageHook(update.Message) {
						return
					}

					NewContext(s, &update).HandleUpdate()
				}()
			}
		}
	}

	util.InitLockerTidy(time.Minute * 30)
	go listen()

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func (s *Server) RegisterChatHook(chatID int64, hook *MessageHook) {
	s.chatHooks.Store(chatID, hook)
}

func (s *Server) RegisterUserHook(userID int64, hook *MessageHook) {
	s.userHooks.Store(userID, hook)
}

func (s *Server) DoMessageHook(m *botapi.Message) (success bool) {
	if m == nil {
		return
	}

	ids := [2]int64{m.Chat.ID, m.From.ID}
	hooks := [2]*MessageHook{nil, nil}
	maps := [2]*sync.Map{&s.chatHooks, &s.userHooks}

	if data, ok := s.chatHooks.Load(m.Chat.ID); ok {
		hooks[0] = data.(*MessageHook)
	}

	if data, ok := s.userHooks.Load(m.From.ID); ok {
		hooks[1] = data.(*MessageHook)
	}

	for i, hook := range hooks {
		if hook == nil {
			continue
		}

		if hook.ExpiresAt.After(time.Now()) {
			go func() {
				if success = hook.Execute(s, m); success {
					maps[i].Delete(ids[i])
				}
			}()
		} else {
			maps[i].Delete(ids[i])
		}
	}

	return
}

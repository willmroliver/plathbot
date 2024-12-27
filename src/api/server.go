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

	messageHooks sync.Map
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

	bot, err := botapi.NewBotAPI("7323800698:AAE2RcvU-g81Iz-nNbRsnglTWmbCuZJJzJA")
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
				if s.DoMessageHook(update.Message) {
					continue
				}

				NewContext(s, &update).HandleUpdate()
			}
		}
	}

	util.InitLockerTidy(time.Minute * 30)
	go listen()

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func (s *Server) RegisterMessageHook(chatID int64, hook *MessageHook) {
	s.messageHooks.Store(chatID, hook)
}

func (s *Server) DoMessageHook(m *botapi.Message) bool {
	if m == nil {
		return false
	}

	data, exists := s.messageHooks.LoadAndDelete(m.Chat.ID)
	if exists && data.(*MessageHook).ExpiresAt.Compare(time.Now()) > -1 {
		data.(*MessageHook).Execute(s, m)
		return true
	}

	return false
}

package api

import (
	"bufio"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	d "github.com/willmroliver/plathbot/src/db"
	"github.com/willmroliver/plathbot/src/util"
	"gorm.io/gorm"
)

var (
	inlineActions  = map[string]InlineAction{}
	commandActions = map[string]CommandAction{}
	callbackAPIs   = map[string]func() *CallbackAPI{}

	onCreateHooks     = []func(*Server){}
	beforeListenHooks = []func(*Server){}
)

type Server struct {
	Bot         *botapi.BotAPI
	DB          *gorm.DB
	CallbackAPI *CallbackAPI
	CommandAPI  *CommandAPI
	InlineAPI   *InlineAPI

	chatHooks sync.Map
	userHooks sync.Map

	callbackAPIs []*CallbackAPI
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
	s := &Server{
		Bot: bot,
		DB:  db,
	}

	s.InlineAPI = &InlineAPI{
		Actions: map[string]InlineAction{},
	}

	s.CommandAPI = &CommandAPI{
		Actions: map[string]CommandAction{},
	}

	s.CallbackAPI = &CallbackAPI{
		Title:   "🚀🌖 P1ath Hub",
		Actions: map[string]CallbackAction{},
		DynamicOptions: func(ctx *Context, cq *botapi.CallbackQuery, cc *CallbackCmd) (opts []map[string]string) {
			apis := s.callbackAPIs

			opts = make([]map[string]string, len(apis))
			public := ctx.Chat.Type != "private"

			for i, a := range apis {
				if a.PrivateOnly && public {
					opts[i] = map[string]string{a.Title: KeyboardLink(ToPrivateString(ctx.Bot, a.Path))}
				} else {
					opts[i] = map[string]string{a.Title: a.Path}
				}
			}

			return
		},
	}

	s.callbackAPIs = make([]*CallbackAPI, 0)

	for cmd, action := range inlineActions {
		s.RegisterInlineAction(cmd, action)
	}

	for cmd, action := range commandActions {
		s.RegisterCommandAction(cmd, action)
	}

	for _, api := range callbackAPIs {
		s.RegisterCallbackAPI(api())
	}

	for _, hook := range onCreateHooks {
		hook(s)
	}

	log.Println("Server: Migrating tables...")
	defer log.Printf("Server: Migration complete")

	d.Migrate(db)

	return s
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

	for _, hook := range beforeListenHooks {
		hook(s)
	}

	go listen()

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func RegisterInlineAction(cmd string, action InlineAction) {
	inlineActions[cmd] = action
}

func RegisterCommandAction(cmd string, action CommandAction) {
	commandActions[cmd] = action
}

func RegisterCallbackAPI(cmd string, api func() *CallbackAPI) {
	callbackAPIs[cmd] = api
}

func (s *Server) RegisterInlineAction(cmd string, action InlineAction) {
	s.InlineAPI.Actions[cmd] = action
}

func (s *Server) RegisterCommandAction(cmd string, action CommandAction) {
	s.CommandAPI.Actions[cmd] = action
}

func (s *Server) RegisterCallbackAPI(api *CallbackAPI) {
	cmd := api.Path
	if i := strings.LastIndex(cmd, "/"); i != -1 {
		cmd = cmd[i+1:]
	}

	s.callbackAPIs = append(s.callbackAPIs, api)
	s.CallbackAPI.Actions[cmd] = api.Select
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

func OnCreate(hook func(*Server)) {
	onCreateHooks = append(onCreateHooks, hook)
}

func BeforeListen(hook func(*Server)) {
	beforeListenHooks = append(beforeListenHooks, hook)
}

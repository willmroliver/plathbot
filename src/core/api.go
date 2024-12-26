package core

import (
	"errors"
	"fmt"
	"log"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/account"
	"github.com/willmroliver/plathbot/src/apis"
	"github.com/willmroliver/plathbot/src/db"
	"github.com/willmroliver/plathbot/src/games"
	"github.com/willmroliver/plathbot/src/server"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	DonateLink string = "https://support.wwf.org.uk/"
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
)

var (
	accountAPI  = account.API()
	gamesAPI    = games.API()
	callbackAPI = &apis.Callback{
		Title: "PlathHub",
		Actions: map[string]apis.CallbackAction{
			"account": accountAPI.Select,
			"games":   gamesAPI.Select,
		},
	}
)

func NewServer() *server.Server {
	conn, err := db.Open("test.db")
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database %q: %q", "test.db", err.Error()))
	}

	s := server.NewServer(conn)
	ConfigureServer(s)

	return s
}

func ConfigureServer(s *server.Server) {
	s.UpdateHandler = func(server *server.Server, update *botapi.Update) {
		NewContext(server, update).HandleUpdate()
	}
}

func handleCommand(s *server.Server, message *botapi.Message, cmd string) error {
	api := map[string]func(*server.Server, *botapi.Chat, *botapi.Message){
		"/start":   sendHelp,
		"/help":    sendHelp,
		"/fact":    sendFact,
		"/account": accountAPI.Expose,
		"/games":   gamesAPI.Expose,
		"/adopt": func(s *server.Server, c *botapi.Chat, m *botapi.Message) {
			if !util.TryLockFor(fmt.Sprintf("%d adopt&donate", c.ID), time.Second*3) {
				util.SendBasic(s.Bot, c.ID, AdoptLink)
			}
		},
		"/donate": func(s *server.Server, c *botapi.Chat, m *botapi.Message) {
			if util.TryLockFor(fmt.Sprintf("%d adopt&donate", c.ID), time.Second*3) {
				util.SendBasic(s.Bot, c.ID, DonateLink)
			}
		},
	}

	if action, exists := api[cmd]; exists {
		action(s, message.Chat, nil)
		return nil
	}

	return errors.New("command does not exist")
}

func handleCallbackQuery(s *server.Server, message *botapi.CallbackQuery, cmd string) {
	go callbackAPI.Select(s, message, apis.NewCallbackCmd(cmd))
}

func handleInlineQuery(s *server.Server, message *botapi.InlineQuery) error {
	api := map[string]func(*server.Server, *botapi.InlineQuery) error{
		"fact":   requestFact,
		"adopt":  requestAdopt,
		"donate": requestDonate,
	}

	if action, exists := api[message.Query]; exists {
		if err := action(s, message); err != nil {
			log.Printf("An error ocurred: %s", err.Error())
			return err
		}

		return nil
	}

	return errors.New("query does not exist")
}

func sendHelp(s *server.Server, c *botapi.Chat, m *botapi.Message) {
	public := fmt.Sprintf(`
	Welcome to the P1athHub - Next stop, the moon 🚀🌖

	Wanna talk? %s
	
	Some things I can do in public chats: try them!
	
	🐾 /plath@help 😣
	🐾 /plath@fact 🧠
	🐾 /plath@adopt 🐼
	🐾 /plath@donate 💸
	🐾 /plath@account 💻
	🐾 /plath@games 🎮

	Telegram won't let me spam group chats, so some of these have rate limits... Sorry!
	`, util.AtBotString(s.Bot))

	private := `
	Hey, it's P1ath 🚀🌖

	What can I help you with?
	
	🐾 /help 😣		- You've made it this far
	🐾 /fact 🧠		- Just for fun :)
	🐾 /adopt 🐼 	- Adopt a platypus
	🐾 /donate 💸	- Support a good cause
	🐾 /account 💻	- Manage your account
	🐾 /games 🎮	- Let's goooo
	`

	text := public
	if c.Type == "private" {
		text = private
	} else if !util.TryLockFor(fmt.Sprintf("%d help", c.ID), time.Second*5) {
		return
	}

	msg := botapi.NewMessage(c.ID, text)
	msg.ParseMode = botapi.ModeMarkdown
	s.Bot.Send(msg)
}

func sendFact(s *server.Server, c *botapi.Chat, m *botapi.Message) {
	if c.Type != "private" && !util.TryLockFor(fmt.Sprintf("%d fact", c.ID), time.Second*5) {
		return
	}

	if _, err := s.Bot.Send(botapi.NewMessage(
		c.ID,
		getFact(),
	)); err != nil {
		log.Printf("Got an error sending a fact: %q", err.Error())
	}
}

func getFact() string {
	facts := []string{
		"The collective noun for 'Platypus' is a 'Pandemonium'.",
		"When European naturalist George Shaw was first presented with a platypus in the 1790s, he thought someone was pulling an elaborate prank.",
		"Bizarrely, platypuses lack a traditional stomach that secretes hydrochloric acid or digestive juices.",
		"Platypuses glow under a blacklight.",
		"A baby platypus is called a 'Puggle'.",
		"Platypuses can sense electrical fields.",
		"Platypuses are one of only two egg-laying mammals.",
		"Platypuses are venomous: Male platypuses have a hollow spur on each hind leg connected to a venom secreting gland.",
		"Platypuses are thought to have evolved from one of Australia's oldest mammals, the Steropodon Galmani",
		"The 20-cent coin in Australia has the image of a platypus on it.",
		"Until the magazine National Geographic published a picture of a platypus in 1939, most of the world had never heard of the platypus.",
		"The platypus will sometimes bury its bill into mud and then wiggle it to attract prey.",
		"The platypus was one of the mascots for the 2000 Summer Olympics held in Sydney, Australia.",
		"One nickname for the platypus is the duck mole because it resembles both of these species.",
		"To date, the oldest platypus fossil found is over 100,000 years old.",
	}

	return facts[util.PseudoRandInt(len(facts), true)]
}

func requestAdopt(s *server.Server, query *botapi.InlineQuery) error {
	return util.RequestBasic(s.Bot, query, "Adopt a Platypus", AdoptLink)
}

func requestDonate(s *server.Server, query *botapi.InlineQuery) error {
	return util.RequestBasic(s.Bot, query, "Donate to WWF", DonateLink)
}

func requestFact(s *server.Server, query *botapi.InlineQuery) (err error) {
	text := "/plath@fact"
	if query.ChatType == "private" {
		text = "fact"
	}

	a := botapi.NewInlineQueryResultArticle(query.ID, "Plath Fact!", text)
	c := botapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{a},
	}

	_, err = s.Bot.Request(c)
	return
}

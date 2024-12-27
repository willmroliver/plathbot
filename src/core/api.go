package core

import (
	"fmt"
	"log"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/account"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/db"
	"github.com/willmroliver/plathbot/src/games"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
	DonateLink string = "https://support.wwf.org.uk/"
)

var (
	accountAPI = account.API()
	gamesAPI   = games.API()

	inlineAPI = &api.InlineAPI{
		Actions: map[string]api.InlineAction{
			"fact":   requestFact,
			"adopt":  requestAdopt,
			"donate": requestDonate,
		},
	}

	commandAPI = &api.CommandAPI{
		Actions: map[string]api.CommandAction{
			"/start": sendHelp,
			"/help":  sendHelp,
			"/fact":  sendFact,
			"/account": func(c *api.Context, m *botapi.Message) {
				accountAPI.Expose(c, nil)
			},
			"/games": func(c *api.Context, m *botapi.Message) {
				gamesAPI.Expose(c, nil)
			},
			"/adopt": func(c *api.Context, m *botapi.Message) {
				if util.TryLockFor(fmt.Sprintf("%d adopt&donate", c.Chat.ID), time.Second*3) {
					util.SendBasic(c.Bot, c.Chat.ID, AdoptLink)
				}
			},
			"/donate": func(c *api.Context, m *botapi.Message) {
				if util.TryLockFor(fmt.Sprintf("%d adopt&donate", c.Chat.ID), time.Second*3) {
					util.SendBasic(c.Bot, c.Chat.ID, DonateLink)
				}
			},
		},
	}

	callbackAPI = &api.CallbackAPI{
		Title: "PlathHub",
		Actions: map[string]api.CallbackAction{
			accountAPI.Path: accountAPI.Select,
			gamesAPI.Path:   gamesAPI.Select,
		},
	}
)

func NewServer() *api.Server {
	conn, err := db.Open("test.db")
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database %q: %q", "test.db", err.Error()))
	}

	s := api.NewServer(conn)

	s.CallbackAPI = callbackAPI
	s.CommandAPI = commandAPI
	s.InlineAPI = inlineAPI

	return s
}

func sendHelp(c *api.Context, m *botapi.Message) {
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
	`, util.AtBotString(c.Bot))

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
	if c.Chat.Type == "private" {
		text = private
	} else if !util.TryLockFor(fmt.Sprintf("%d help", c.Chat.ID), time.Second*5) {
		return
	}

	msg := botapi.NewMessage(c.Chat.ID, text)
	msg.ParseMode = botapi.ModeMarkdown
	c.Bot.Send(msg)
}

func sendFact(c *api.Context, m *botapi.Message) {
	if c.Chat.Type != "private" && !util.TryLockFor(fmt.Sprintf("%d fact", c.Chat.ID), time.Second*5) {
		return
	}

	if _, err := c.Bot.Send(botapi.NewMessage(
		c.Chat.ID,
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

func requestAdopt(c *api.Context, query *botapi.InlineQuery) {
	util.RequestBasic(c.Bot, query, "Adopt a Platypus", AdoptLink)
}

func requestDonate(c *api.Context, query *botapi.InlineQuery) {
	util.RequestBasic(c.Bot, query, "Donate to WWF", DonateLink)
}

func requestFact(c *api.Context, query *botapi.InlineQuery) {
	text := "/plath@fact"
	if query.ChatType == "private" {
		text = "fact"
	}

	c.Bot.Request(botapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{botapi.NewInlineQueryResultArticle(query.ID, "Plath Fact!", text)},
	})
}

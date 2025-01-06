package core

import (
	"fmt"
	"log"
	"os"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	account "github.com/willmroliver/plathbot/src/api_account"
	emoji "github.com/willmroliver/plathbot/src/api_emoji"
	games "github.com/willmroliver/plathbot/src/api_games"
	pfp "github.com/willmroliver/plathbot/src/api_pfp"
	stats "github.com/willmroliver/plathbot/src/api_stats"
	"github.com/willmroliver/plathbot/src/db"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
	DonateLink string = "https://support.wwf.org.uk/"
)

var (
	accountAPI = account.API()
	gamesAPI   = games.API()
	emojiAPI   = emoji.API()
	statsAPI   = stats.API()

	inlineAPI = &api.InlineAPI{
		Actions: map[string]api.InlineAction{
			"fact":   requestFact,
			"adopt":  requestAdopt,
			"donate": requestDonate,
		},
	}

	commandAPI = &api.CommandAPI{
		Actions: map[string]api.CommandAction{
			"/start": sendStart,
			"/help":  sendHelp,
			"/fact":  sendFact,
			"/pfp":   pfp.API,
			"/hub": func(c *api.Context, m *botapi.Message, args ...string) {
				callbackAPI.Expose(c, nil, nil)
			},
			"/adopt": func(c *api.Context, m *botapi.Message, args ...string) {
				if util.TryLockFor(fmt.Sprintf("%d adopt&donate", c.Chat.ID), time.Second*3) {
					api.SendBasic(c.Bot, c.Chat.ID, AdoptLink)
				}
			},
			"/donate": func(c *api.Context, m *botapi.Message, args ...string) {
				if util.TryLockFor(fmt.Sprintf("%d adopt&donate", c.Chat.ID), time.Second*3) {
					api.SendBasic(c.Bot, c.Chat.ID, DonateLink)
				}
			},
		},
	}

	callbackAPI = &api.CallbackAPI{
		Title: "🚀🌖 P1ath Hub",
		Actions: map[string]api.CallbackAction{
			accountAPI.Path: accountAPI.Select,
			gamesAPI.Path:   gamesAPI.Select,
			emojiAPI.Path:   emojiAPI.Select,
			statsAPI.Path:   statsAPI.Select,
		},
		DynamicOptions: func(ctx *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) (opts []map[string]string) {
			apis := []*api.CallbackAPI{
				accountAPI,
				gamesAPI,
				emojiAPI,
				statsAPI,
			}

			opts = make([]map[string]string, len(apis))
			public := ctx.Chat.Type != "private"

			for i, a := range apis {
				if a.PrivateOnly && public {
					opts[i] = map[string]string{a.Title: api.KeyboardLink(api.ToPrivateString(ctx.Bot, a.Path))}
				} else {
					opts[i] = map[string]string{a.Title: a.Path}
				}
			}

			return
		},
	}
)

func NewServer() *api.Server {
	conn, err := db.Open(os.Getenv("MOUNT_DIR") + "/" + os.Getenv("DB_NAME"))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database %q: %q", "test.db", err.Error()))
	}

	s := api.NewServer(conn)

	s.CallbackAPI = callbackAPI
	s.CommandAPI = commandAPI
	s.InlineAPI = inlineAPI

	return s
}

func sendStart(c *api.Context, m *botapi.Message, args ...string) {
	if len(args) < 2 {
		sendHelp(c, m)
		return
	}

	c.Server.CommandAPI.Actions["/"+args[1]](c, m)
}

func sendHelp(c *api.Context, m *botapi.Message, args ...string) {
	public := fmt.Sprintf(`
	Welcome to the P1athHub - Next stop, the moon 🚀🌖

	Wanna talk? %s
	
	🐾 /hub 🚀   	
	🐾 /help 😣		
	🐾 /fact 🧠		
	🐾 /adopt 🐼 	
	🐾 /donate 💸	
	🐾 /account 💻	
	🐾 /games 🎮	
	🐾 /emojis 🙂	
	🐾 /stats 📊	

	For more info, pass 'help' or '?' to any of these commands:
		"/games help"
	`, api.AtBotString(c.Bot))

	private := `
	Hey, it's P1ath 🚀🌖

	What can I help you with?
	
	🐾 /hub 🚀   	- We all prefer buttons
	🐾 /help 😣		- You've made it this far
	🐾 /fact 🧠		- Just for fun :)
	🐾 /adopt 🐼 	- Adopt a platypus
	🐾 /donate 💸	- Support a good cause
	🐾 /account 💻	- Manage your account
	🐾 /games 🎮	- Let's goooo
	🐾 /emojis 🙂	- React leaderboards <3
	🐾 /stats 📊	- Top the charts for airdrops!

	For more info, pass 'help' or '?' to any of these commands:
		"/games help"
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

func sendFact(c *api.Context, m *botapi.Message, args ...string) {
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
	api.RequestBasic(c.Bot, query, "Adopt a Platypus", AdoptLink)
}

func requestDonate(c *api.Context, query *botapi.InlineQuery) {
	api.RequestBasic(c.Bot, query, "Donate to WWF", DonateLink)
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

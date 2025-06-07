package core

import (
	"fmt"
	"log"
	"os"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/db"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
	DonateLink string = "https://support.wwf.org.uk/"
)

func NewServer() *api.Server {
	dbn := os.Getenv("MOUNT_DIR") + "/" + os.Getenv("DB_NAME")

	conn, err := db.Open(dbn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database %q: %q", dbn, err.Error()))
	}

	log.Printf("Server: Connection opened to %s\n", dbn)

	s := api.NewServer(conn)

	s.RegisterInlineAction("fact", requestFact)
	s.RegisterInlineAction("adopt", requestAdopt)
	s.RegisterInlineAction("donate", requestDonate)

	s.RegisterCommandAction("/start", func(c *api.Context, m *botapi.Message, args ...string) {
		s.CallbackAPI.Expose(c, nil, nil)
	})
	s.RegisterCommandAction("/help", func(ctx *api.Context, m *botapi.Message, s ...string) {
		ctx.Server.CallbackAPI.SendHelp(ctx, nil, nil)
	})
	s.RegisterCommandAction("/hub", func(c *api.Context, m *botapi.Message, args ...string) {
		s.CallbackAPI.Expose(c, nil, nil)
	})

	s.RegisterCommandAction("/fact", sendFact)

	s.RegisterCommandAction("/adopt", func(c *api.Context, m *botapi.Message, args ...string) {
		if util.TryLockFor(fmt.Sprintf("%d adopt&donate", c.Chat.ID), time.Second*3) {
			api.SendBasic(c.Bot, c.Chat.ID, AdoptLink)
		}
	})
	s.RegisterCommandAction("/donate", func(c *api.Context, m *botapi.Message, args ...string) {
		if util.TryLockFor(fmt.Sprintf("%d adopt&donate", c.Chat.ID), time.Second*3) {
			api.SendBasic(c.Bot, c.Chat.ID, DonateLink)
		}
	})

	return s
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

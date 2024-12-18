package io

import (
	"errors"
	"log"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/games"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	DonateLink string = "https://support.wwf.org.uk/"
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
)

type BotContext struct {
	bot    *botapi.BotAPI
	update botapi.Update
}

func NewBotContext(bot *botapi.BotAPI, update botapi.Update) *BotContext {
	return &BotContext{
		bot,
		update,
	}
}

func (ctx *BotContext) HandleUpdate() {
	ctx.HandleMessage()
	ctx.HandleCallbackQuery()
	ctx.HandleInlineQuery()
}

func (ctx *BotContext) HandleMessage() {
	m := ctx.update.Message
	if m == nil {
		return
	}

	user := m.From
	text := m.Text

	if user == nil {
		return
	}

	log.Printf("Received %q from %s", text, user.FirstName)

	var err error

	switch {
	case strings.HasPrefix(text, "/"):
		err = ctx.handleCommand(m, text)
	default:
		break
	}

	if err != nil {
		log.Printf("An error ocurred: %s", err.Error())
	}
}

func (ctx *BotContext) HandleCallbackQuery() {
	m := ctx.update.CallbackQuery
	if m == nil {
		return
	}

	var api = map[string]func(*botapi.BotAPI, *botapi.CallbackQuery, string){
		"games": games.HandleCallbackQuery,
	}

	cmd := m.Data

	for key, action := range api {
		if strings.HasPrefix(cmd, key+"/") {
			go action(ctx.bot, m, cmd[len(key)+1:])
			break
		}
	}
}

func (ctx *BotContext) HandleInlineQuery() {
	m := ctx.update.InlineQuery
	if m == nil {
		return
	}

	api := map[string]func(*BotContext, *botapi.InlineQuery) error{
		"fact":   requestFact,
		"adopt":  requestAdopt,
		"donate": requestDonate,
	}

	action, exists := api[m.Query]

	if exists {
		err := action(ctx, m)
		if err != nil {
			log.Printf("An error ocurred: %s", err.Error())
		}
	}
}

func (ctx *BotContext) handleCommand(message *botapi.Message, cmd string) error {
	api := map[string]func(*botapi.BotAPI, *botapi.Message) error{
		"/start": func(bot *botapi.BotAPI, m *botapi.Message) (err error) {
			return util.SendBasic(bot, m.Chat.ID, "Hi, I'm P1ath, your fav3333 crypto platypus :)")
		},
		"/adopt": func(bot *botapi.BotAPI, m *botapi.Message) (err error) {
			return util.SendBasic(bot, m.Chat.ID, AdoptLink)
		},
		"/donate": func(bot *botapi.BotAPI, m *botapi.Message) (err error) {
			return util.SendBasic(bot, m.Chat.ID, DonateLink)
		},
		"/help":  sendHelp,
		"/fact":  sendFact,
		"/games": games.SendMenu,
	}

	action, exists := api[cmd]

	if exists {
		return action(ctx.bot, message)
	}

	return errors.New("action does not exist")
}

func sendHelp(bot *botapi.BotAPI, m *botapi.Message) (err error) {
	return util.SendBasic(bot, m.Chat.ID, `
	Plath commands currently available: try them!
	
	/start	
	/help
	/fact 				- Just for fun :)
	/adopt 				- Adopt a platypus
	/donate				- Support a good cause
	`)
}

func sendFact(bot *botapi.BotAPI, m *botapi.Message) (err error) {
	_, err = bot.Send(botapi.NewMessage(
		m.Chat.ID,
		getFact(),
	))

	if err != nil {
		log.Printf("Got an error sending a fact: %q", err.Error())
	}

	return
}

func requestAdopt(ctx *BotContext, query *botapi.InlineQuery) error {
	return util.RequestBasic(ctx.bot, query, "Adopt a Platypus", AdoptLink)
}

func requestDonate(ctx *BotContext, query *botapi.InlineQuery) error {
	return util.RequestBasic(ctx.bot, query, "Donate to WWF", DonateLink)
}

func requestFact(ctx *BotContext, query *botapi.InlineQuery) (err error) {
	a := botapi.NewInlineQueryResultArticle(query.ID, "Plath Fact!", "/fact")
	c := botapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{a},
	}

	_, err = ctx.bot.Request(c)
	return
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

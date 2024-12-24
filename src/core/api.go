package core

import (
	"errors"
	"fmt"
	"log"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/account"
	"github.com/willmroliver/plathbot/src/apis"
	"github.com/willmroliver/plathbot/src/games"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	DonateLink string = "https://support.wwf.org.uk/"
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
)

func HandleCommand(bot *botapi.BotAPI, message *botapi.Message, cmd string) error {
	api := map[string]func(*botapi.BotAPI, *botapi.Message) error{
		"/start":   sendHelp,
		"/help":    sendHelp,
		"/fact":    sendFact,
		"/account": account.SendOptions,
		"/games":   games.SendOptions,
		"/adopt": func(bot *botapi.BotAPI, m *botapi.Message) (err error) {
			if !util.TryLockFor(fmt.Sprintf("%d adopt&donate", m.Chat.ID), time.Second*3) {
				return nil
			}

			return util.SendBasic(bot, m.Chat.ID, AdoptLink)
		},
		"/donate": func(bot *botapi.BotAPI, m *botapi.Message) (err error) {
			if !util.TryLockFor(fmt.Sprintf("%d adopt&donate", m.Chat.ID), time.Second*3) {
				return nil
			}

			return util.SendBasic(bot, m.Chat.ID, DonateLink)
		},
	}

	if action, exists := api[cmd]; exists {
		return action(bot, message)
	}

	return errors.New("command does not exist")
}

func HandleCallbackQuery(bot *botapi.BotAPI, message *botapi.CallbackQuery, cmd string) {
	api := apis.Callback{
		"account": account.HandleCallbackQuery,
		"games":   games.HandleCallbackQuery,
	}

	go api.Next(bot, message, apis.NewCallbackCmd(cmd))
}

func HandleInlineQuery(bot *botapi.BotAPI, message *botapi.InlineQuery) error {
	api := map[string]func(*botapi.BotAPI, *botapi.InlineQuery) error{
		"fact":   requestFact,
		"adopt":  requestAdopt,
		"donate": requestDonate,
	}

	if action, exists := api[message.Query]; exists {
		if err := action(bot, message); err != nil {
			log.Printf("An error ocurred: %s", err.Error())
			return err
		}

		return nil
	}

	return errors.New("query does not exist")
}

func sendHelp(bot *botapi.BotAPI, m *botapi.Message) (err error) {
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
	`, util.AtBotString(bot))

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
	if m.Chat.Type == "private" {
		text = private
	} else if !util.TryLockFor(fmt.Sprintf("%d help", m.Chat.ID), time.Second*5) {
		return
	}

	msg := botapi.NewMessage(m.Chat.ID, text)
	msg.ParseMode = botapi.ModeMarkdown

	_, err = bot.Send(msg)
	return
}

func sendFact(bot *botapi.BotAPI, m *botapi.Message) (err error) {
	if m.Chat.Type != "private" && !util.TryLockFor(fmt.Sprintf("%d fact", m.Chat.ID), time.Second*5) {
		return nil
	}

	_, err = bot.Send(botapi.NewMessage(
		m.Chat.ID,
		getFact(),
	))

	if err != nil {
		log.Printf("Got an error sending a fact: %q", err.Error())
	}

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

func requestAdopt(bot *botapi.BotAPI, query *botapi.InlineQuery) error {
	return util.RequestBasic(bot, query, "Adopt a Platypus", AdoptLink)
}

func requestDonate(bot *botapi.BotAPI, query *botapi.InlineQuery) error {
	return util.RequestBasic(bot, query, "Donate to WWF", DonateLink)
}

func requestFact(bot *botapi.BotAPI, query *botapi.InlineQuery) (err error) {
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

	_, err = bot.Request(c)
	return
}

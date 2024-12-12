package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"os"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/games"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	DonateLink string = "https://support.wwf.org.uk/"
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
)

var (
	bot *botapi.BotAPI
)

func main() {
	var err error
	bot, err = botapi.NewBotAPI("7323800698:AAE2RcvU-g81Iz-nNbRsnglTWmbCuZJJzJA")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	u := botapi.NewUpdate(0)
	u.Timeout = 60

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	ch := bot.GetUpdatesChan(u)

	go receiveUpdates(ctx, ch)

	log.Println("Listening for updates. Press enter to exit")

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func receiveUpdates(ctx context.Context, updates botapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			handleUpdate(update)
		}
	}
}

func handleUpdate(u botapi.Update) {
	switch {
	case u.Message != nil:
		handleMessage(u.Message)
	case u.CallbackQuery != nil:
		handleCallbackQuery(u.CallbackQuery)
	case u.InlineQuery != nil:
		handleInlineQuery(u.InlineQuery)
	default:
		break
	}
}

func handleMessage(m *botapi.Message) {
	user := m.From
	text := m.Text

	if user == nil {
		return
	}

	log.Printf("Received %q from %s", text, user.FirstName)

	var err error

	switch {
	case strings.HasPrefix(text, "/"):
		err = handleCommand(bot, m.Chat.ID, text)
	default:
		break
	}

	if err != nil {
		log.Printf("An error ocurred: %s", err.Error())
	}
}

func handleCallbackQuery(m *botapi.CallbackQuery) {
	var api = map[string]func(*botapi.BotAPI, *botapi.CallbackQuery, string){
		"games": games.HandleCallbackQuery,
	}

	cmd := m.Data

	for key, action := range api {
		if strings.HasPrefix(cmd, key+"/") {
			action(bot, m, cmd[len(key)+1:])
			break
		}
	}
}

func handleInlineQuery(m *botapi.InlineQuery) {
	api := map[string]func(*botapi.BotAPI, *botapi.InlineQuery) error{
		"fact":   requestFact,
		"adopt":  requestAdopt,
		"donate": requestDonate,
	}

	log.Printf("Query %q from %s", m.Query, m.From.FirstName)

	action, exists := api[m.Query]

	if exists {
		err := action(bot, m)
		if err != nil {
			log.Printf("An error ocurred: %s", err.Error())
		}
	}
}

func handleCommand(bot *botapi.BotAPI, chatID int64, cmd string) error {
	api := map[string]func(*botapi.BotAPI, int64) error{
		"/start": func(bot *botapi.BotAPI, chatID int64) (err error) {
			return util.SendBasic(bot, chatID, "Hi, I'm P1ath, your fav3333 crypto platypus :)")
		},
		"/adopt": func(bot *botapi.BotAPI, chatID int64) (err error) {
			return util.SendBasic(bot, chatID, AdoptLink)
		},
		"/donate": func(bot *botapi.BotAPI, chatID int64) (err error) {
			return util.SendBasic(bot, chatID, DonateLink)
		},
		"/help":  sendHelp,
		"/fact":  sendFact,
		"/games": games.SendMenu,
	}

	action, exists := api[cmd]

	if exists {
		return action(bot, chatID)
	}

	return errors.New("action does not exist")
}

func sendHelp(bot *botapi.BotAPI, chatID int64) (err error) {
	return util.SendBasic(bot, chatID, `
	Plath commands currently available: try them!
	
	/start	
	/help
	/fact 				- Just for fun :)
	/adopt 				- Adopt a platypus
	/donate				- Support a good cause
	`)
}

func sendFact(bot *botapi.BotAPI, chatID int64) (err error) {
	_, err = bot.Send(botapi.NewMessage(
		chatID,
		getFact(),
	))

	return
}

func requestAdopt(bot *botapi.BotAPI, query *botapi.InlineQuery) error {
	return util.RequestBasic(bot, query, "Adopt a Platypus", AdoptLink)
}

func requestDonate(bot *botapi.BotAPI, query *botapi.InlineQuery) error {
	return util.RequestBasic(bot, query, "Donate to WWF", DonateLink)
}

func requestFact(bot *botapi.BotAPI, query *botapi.InlineQuery) (err error) {
	a := botapi.NewInlineQueryResultArticle(query.ID, "Plath Fact!", "/fact")
	c := botapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{a},
	}

	_, err = bot.Request(c)
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
	}

	return facts[util.PseudoRandInt(len(facts), true)]
}

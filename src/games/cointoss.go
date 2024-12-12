package games

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/util"
)

type CoinToss struct {
	ID       int64
	bot      *botapi.BotAPI
	chatID   int64
	players  []*botapi.User
	chooses  int
	nextMove string
}

var running = map[int64]*CoinToss{}

func CointossQuery(bot *botapi.BotAPI, query *botapi.CallbackQuery, cmd string) {
	if cmd == "" {
		_, exists := running[query.From.ID]
		if !exists {
			game := NewCoinToss(bot, query.Message.Chat.ID, query.From)

			if game.RequestToss() == nil {
				running[game.ID] = game
			}
		}

		return
	}

	parts := strings.Split(cmd, "/")
	cmd = parts[0]
	id, err := strconv.ParseInt(parts[1], 10, 64)

	if err != nil {
		log.Printf("Invalid game ID: %q", err.Error())
		return
	}

	game, exists := running[id]

	if !exists {
		log.Printf("Game %d does not exist", id)
		return
	}

	switch cmd {
	case "accept":
		if game.AcceptToss(query.From) != nil {
			delete(running, id)
		}
	case "heads":
		game.Toss(true)
	case "tails":
		game.Toss(false)
	default:
		return
	}
}

func NewCoinToss(bot *botapi.BotAPI, chatID int64, player *botapi.User) *CoinToss {
	return &CoinToss{
		ID:       player.ID,
		bot:      bot,
		chatID:   chatID,
		players:  []*botapi.User{player, nil},
		chooses:  -1,
		nextMove: "request",
	}
}

func (ct *CoinToss) GetChosen() *botapi.User {
	if ct.chooses == -1 {
		ct.chooses = util.PseudoRandInt(2, true)
	}

	return ct.players[ct.chooses]
}

func (ct *CoinToss) RequestToss() (err error) {
	if ct.nextMove != "request" {
		return
	}

	msg := ct.newMessage(util.AtUser(ct.players[0]) + " wants to toss a coin...")
	msg.ReplyMarkup = botapi.NewInlineKeyboardMarkup(
		botapi.NewInlineKeyboardRow(
			botapi.NewInlineKeyboardButtonData("Play!", ct.getButtonData("accept")),
		),
	)

	_, err = ct.bot.Send(msg)

	if err == nil {
		ct.nextMove = "accept"
	}

	return
}

func (ct *CoinToss) AcceptToss(player *botapi.User) (err error) {
	if ct.nextMove != "accept" {
		return
	}

	ct.players[1] = player

	msg := ct.newMessage(fmt.Sprintf("%s, heads or tails?", util.AtUser(ct.GetChosen())))
	msg.ReplyMarkup = botapi.NewInlineKeyboardMarkup(
		botapi.NewInlineKeyboardRow(
			botapi.NewInlineKeyboardButtonData("Heads", ct.getButtonData("heads")),
			botapi.NewInlineKeyboardButtonData("Tails", ct.getButtonData("tails")),
		),
	)

	_, err = ct.bot.Send(msg)

	if err == nil {
		ct.nextMove = "toss"
	}

	return
}

func (ct *CoinToss) Toss(heads bool) {
	defer func() {
		delete(running, ct.ID)
	}()

	if ct.nextMove != "toss" {
		return
	}

	choice := "tails"
	if heads {
		choice = "heads"
	}

	var err error

	_, err = ct.bot.Send(ct.newMessage(fmt.Sprintf("%s chooses %q ...", util.AtUser(ct.GetChosen()), choice)))
	if err != nil {
		return
	}

	heads = util.PseudoRandInt(2, false) == 1
	result := "tails"
	if heads {
		result = "heads"
	}

	_, err = ct.bot.Send(ct.newMessage(fmt.Sprintf("%s The coin lands... %q", ct.playerPrefix(), result)))
	if err != nil {
		return
	}

	winner := ct.players[0]
	if (choice == result && ct.chooses != 0) || (choice != result && ct.chooses == 0) {
		winner = ct.players[1]
	}

	ct.bot.Send(ct.newMessage(fmt.Sprintf(
		"%s The winner is %s!",
		ct.playerPrefix(),
		util.AtUser(winner),
	)))
}

func (ct *CoinToss) getButtonData(cmd string) string {
	log.Printf("Data: %q", "games/cointoss/"+cmd+"/"+strconv.FormatInt(ct.ID, 10))
	return "games/cointoss/" + cmd + "/" + strconv.FormatInt(ct.ID, 10)
}

func (ct *CoinToss) newMessage(text string) botapi.MessageConfig {
	msg := botapi.NewMessage(ct.chatID, text)
	msg.ParseMode = "Markdown"
	return msg
}

func (ct *CoinToss) playerPrefix() string {
	return fmt.Sprintf("%s vs %s:", util.AtUser(ct.players[0]), util.AtUser(ct.players[1]))
}

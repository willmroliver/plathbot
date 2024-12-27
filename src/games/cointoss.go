package games

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/util"
)

type CoinToss struct {
	*api.Interaction[string]
	ID      int64
	bot     *botapi.BotAPI
	players []*botapi.User
	chooses int
	mu      sync.Mutex
	updated time.Time
}

var running = sync.Map{}

func CointossQuery(c *api.Context, query *botapi.CallbackQuery, cmd *api.CallbackCmd) {
	running.Range(func(key any, value any) bool {
		game := value.(*CoinToss)
		if game.updated.Add(time.Minute*5).Compare(time.Now()) == -1 {
			running.Delete(key)
		}

		return true
	})

	action := cmd.Get()

	if action == "" {
		_, exists := running.Load(query.From.ID)
		if !exists {
			game := NewCoinToss(c.Bot, query.Message, query.From)

			if game.RequestToss(query) == nil {
				running.Store(game.ID, game)
			}
		}

		return
	}

	id, err := strconv.ParseInt(cmd.Next().Get(), 10, 64)
	if err != nil {
		log.Printf("Invalid game ID: %q", err.Error())
		return
	}

	val, exists := running.Load(id)
	if !exists {
		log.Printf("Game %d does not exist", id)
		return
	}

	game := val.(*CoinToss)

	if !game.mu.TryLock() {
		return
	}

	defer game.mu.Unlock()

	switch action {
	case "accept":
		if game.AcceptToss(query) != nil {
			running.Delete(game.ID)
		}
	case "heads", "tails":
		if query.From.ID != game.GetChosen().ID {
			break
		}

		game.Toss(query, action == "heads")
	default:
		break
	}

	// Throttle a bit to protect message rate limit
	// Works well with the mutex lock strategy above to stagger games
	time.Sleep(time.Millisecond * 500)
}

func NewCoinToss(bot *botapi.BotAPI, message *botapi.Message, player *botapi.User) *CoinToss {
	return &CoinToss{
		Interaction: api.NewInteraction[string](message, "request"),
		ID:          player.ID,
		bot:         bot,
		players:     []*botapi.User{player, nil},
		chooses:     -1,
		updated:     time.Now(),
	}
}

func (ct *CoinToss) GetChosen() *botapi.User {
	if ct.chooses == -1 {
		ct.chooses = util.PseudoRandInt(2, true)
	}

	return ct.players[ct.chooses]
}

func (ct *CoinToss) RequestToss(query *botapi.CallbackQuery) (err error) {
	if !ct.Is("request") {
		return
	}

	msg := ct.NewMessageUpdate(
		util.AtUserString(ct.players[0])+" wants to toss a coin...",
		util.InlineKeyboard([]map[string]string{{
			"Play!": ct.getCmd("accept"),
		}}),
	)

	if _, err = ct.bot.Send(msg); err != nil {
		log.Printf("Error in RequestToss(): %q", err.Error())
		return
	}

	ct.Mutate("accept", query.Message)
	return
}

func (ct *CoinToss) AcceptToss(query *botapi.CallbackQuery) (err error) {
	if !ct.Is("accept") {
		return
	}

	ct.players[1] = query.From

	msg := ct.NewMessageUpdate(
		fmt.Sprintf("%s, heads or tails?", util.AtUserString(ct.GetChosen())),
		util.InlineKeyboard([]map[string]string{{
			"🙉 Heads": ct.getCmd("heads"),
			"🐒 Tails": ct.getCmd("tails"),
		}}),
	)

	if _, err = ct.bot.Send(msg); err != nil {
		log.Printf("Error in AcceptToss(): %q", err.Error())
		return
	}

	ct.Mutate("toss", query.Message)
	return
}

func (ct *CoinToss) Toss(query *botapi.CallbackQuery, heads bool) (err error) {
	defer func() {
		running.Delete(ct.ID)
	}()

	if !ct.Is("toss") {
		return
	}

	choice := "tails"
	if heads {
		choice = "heads"
	}

	msg := ct.NewMessage(fmt.Sprintf("%s chooses %q ...", util.AtUserString(ct.GetChosen()), choice))

	if _, err = ct.bot.Send(msg); err != nil {
		log.Printf("Error in Toss(): %q", err.Error())
		return
	}

	heads = util.PseudoRandInt(2, false) == 1
	result := "tails"
	if heads {
		result = "heads"
	}

	winner := ct.players[0]
	if (choice == result && ct.chooses != 0) || (choice != result && ct.chooses == 0) {
		winner = ct.players[1]
	}

	msg = ct.NewMessage(fmt.Sprintf(`
		%s The coin lands... %q 

		The winner is %s!
	`, ct.playerPrefix(), result, util.AtUserString(winner)))

	if _, err = ct.bot.Send(msg); err != nil {
		log.Printf("Error in Toss(): %q", err.Error())
		return
	}

	return
}

func (ct *CoinToss) getCmd(cmd string) string {
	return fmt.Sprintf("%s/cointoss/%s/%d", Path, cmd, ct.ID)
}

func (ct *CoinToss) playerPrefix() string {
	return fmt.Sprintf("%s vs %s:", util.AtUserString(ct.players[0]), util.AtUserString(ct.players[1]))
}

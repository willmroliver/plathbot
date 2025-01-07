package games

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/service"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	CointossTitle = "🪙 Cointoss"
	CointossPath  = Path + "/cointoss"
)

type CoinToss struct {
	*api.Interaction[string]
	ID      int64
	Bot     *botapi.BotAPI
	Players []*botapi.User
	Chooses int
	Mu      sync.Mutex
}

var cointossRunning = sync.Map{}

func CointossQuery(c *api.Context, query *botapi.CallbackQuery, cmd *api.CallbackCmd) {
	cointossRunning.Range(func(key any, value any) bool {
		game := value.(*CoinToss)
		if game.Age() > time.Minute*5 {
			cointossRunning.Delete(key)
		}

		return true
	})

	action := cmd.Get()

	if action == "" {
		_, exists := cointossRunning.Load(query.From.ID)
		if !exists {
			game := NewCoinToss(c.Bot, query.Message, query.From)

			if game.RequestToss(query) == nil {
				cointossRunning.Store(game.ID, game)
			}
		}

		return
	}

	id, err := strconv.ParseInt(cmd.Next().Get(), 10, 64)
	if err != nil {
		log.Printf("Invalid game ID: %q", err.Error())
		return
	}

	val, exists := cointossRunning.Load(id)
	if !exists {
		log.Printf("Game %d does not exist", id)
		return
	}

	game := val.(*CoinToss)

	if !game.Mu.TryLock() {
		return
	}

	defer game.Mu.Unlock()

	switch action {
	case "accept":
		if game.AcceptToss(query) != nil {
			cointossRunning.Delete(game.ID)
		}
	case "heads", "tails":
		if query.From.ID != game.GetChosen().ID {
			break
		}

		game.Toss(c, query, action == "heads")
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
		Bot:         bot,
		Players:     []*botapi.User{player, nil},
		Chooses:     -1,
	}
}

func (ct *CoinToss) GetChosen() *botapi.User {
	if ct.Chooses == -1 {
		ct.Chooses = util.PseudoRandInt(2, true)
	}

	return ct.Players[ct.Chooses]
}

func (ct *CoinToss) RequestToss(query *botapi.CallbackQuery) (err error) {
	if !ct.Is("request") {
		return
	}

	msg := ct.NewMessageUpdate(
		api.AtUserString(ct.Players[0])+" wants to toss a coin...",
		api.InlineKeyboard([]map[string]string{{"Play!": ct.getCmd("accept")}}),
	)

	if _, err = ct.Bot.Send(msg); err != nil {
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

	ct.Players[1] = query.From

	msg := ct.NewMessageUpdate(
		api.AtUserString(ct.GetChosen())+", heads or tails?",
		api.InlineKeyboard([]map[string]string{{
			"🙉 Heads": ct.getCmd("heads"),
			"🐒 Tails": ct.getCmd("tails"),
		}}),
	)

	if _, err = ct.Bot.Send(msg); err != nil {
		log.Printf("Error in AcceptToss(): %q", err.Error())
		return
	}

	ct.Mutate("toss", query.Message)
	return
}

func (ct *CoinToss) Toss(c *api.Context, query *botapi.CallbackQuery, heads bool) (err error) {
	defer func() {
		cointossRunning.Delete(ct.ID)
	}()

	if !ct.Is("toss") {
		return
	}

	choice := "🐒"
	if heads {
		choice = "🙉"
	}

	gameText := fmt.Sprintf(`
%s: %s
%s chooses %s ...`, CointossTitle, ct.playerPrefix(), api.AtUserString(ct.GetChosen()), choice)

	msg := ct.NewMessageUpdate(gameText, nil)
	ct.Bot.Send(msg)
	time.Sleep(time.Millisecond * 500)

	heads = util.PseudoRandInt(2, false) == 1
	result := "🐒"
	if heads {
		result = "🙉"
	}

	winner := ct.Players[0]
	if (choice == result && ct.Chooses != 0) || (choice != result && ct.Chooses == 0) {
		winner = ct.Players[1]
	}

	xpText := ""
	if ct.Players[0].ID != ct.Players[1].ID {
		var xp int64 = 100

		service.
			NewUserXPService(c.Server.DB).
			UpdateXPs(c.User, service.XPTitleGames, xp)

		xpText = fmt.Sprintf(" +%d XP", xp)
	}

	gameText = fmt.Sprintf(`
%s

The coin lands... %s`, gameText, result)

	msg = ct.NewMessageUpdate(gameText, nil)
	ct.Bot.Send(msg)
	time.Sleep(time.Millisecond * 500)

	gameText = fmt.Sprintf(`
%s

%s wins!%s`, gameText, api.AtUserString(winner), xpText)

	msg = ct.NewMessageUpdate(gameText, nil)
	ct.Bot.Send(msg)

	return
}

func (ct *CoinToss) getCmd(cmd string) string {
	return fmt.Sprintf("%s/%s/%d", CointossPath, cmd, ct.ID)
}

func (ct *CoinToss) playerPrefix() string {
	return api.AtUserString(ct.Players[0]) + " vs " + api.AtUserString(ct.Players[1])
}

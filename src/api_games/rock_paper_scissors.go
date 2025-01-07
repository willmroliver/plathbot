package games

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/service"
)

type Move string

const (
	RockPaperScissorsTitle = "🪨 Rock, 📜 Paper, ✂️ Scissors"
	RockPaperScissorsPath  = Path + "/rockpaperscissors"

	MoveRock     Move = "🪨"
	MovePaper    Move = "📜"
	MoveScissors Move = "✂️"
)

type RockPaperScissors struct {
	*api.Interaction[string]
	ID          int64
	TotalRounds int
	Round       int
	Moves       [][2]Move
	Bot         *botapi.BotAPI
	Players     [2]*botapi.User
	Mu          sync.Mutex
}

var rpsRunning = sync.Map{}

func RockPaperScissorsQuery(c *api.Context, query *botapi.CallbackQuery, cmd *api.CallbackCmd) {
	rpsRunning.Range(func(key any, value any) bool {
		game := value.(*RockPaperScissors)
		if game.Age() > time.Minute*5 {
			rpsRunning.Delete(key)
		}

		return true
	})

	action := cmd.Get()

	if action == "" {
		_, exists := rpsRunning.Load(query.From.ID)
		if !exists {
			game := NewRockPaperScissors(c, 3)

			if game.RequestGame(query) == nil {
				rpsRunning.Store(game.ID, game)
			}
		}

		return
	}

	id, err := strconv.ParseInt(cmd.Next().Get(), 10, 64)
	if err != nil {
		log.Printf("Invalid game ID: %q", err.Error())
		return
	}

	val, exists := rpsRunning.Load(id)
	if !exists {
		log.Printf("Game %d does not exist", id)
		return
	}

	game := val.(*RockPaperScissors)

	if !game.Mu.TryLock() {
		return
	}

	defer game.Mu.Unlock()

	switch action {
	case "accept":
		if game.AcceptGame(c, query) != nil {
			rpsRunning.Delete(game.ID)
		}
	case string(MoveRock), string(MovePaper), string(MoveScissors):
		if i, err := strconv.Atoi(cmd.Next().Get()); err == nil {
			if done, err := game.DoMove(Move(action), i, query); err == nil && done {
				time.Sleep(time.Millisecond * 500)
				game.SendRound(c, query)
			}
		}
	default:
		break
	}
}

func NewRockPaperScissors(c *api.Context, rounds int) *RockPaperScissors {
	return &RockPaperScissors{
		Interaction: api.NewInteraction[string](c.Message, "request"),
		ID:          c.User.ID,
		TotalRounds: rounds,
		Round:       0,
		Moves:       make([][2]Move, rounds),
		Bot:         c.Bot,
		Players:     [2]*botapi.User{c.User, nil},
	}
}

func (g *RockPaperScissors) RequestGame(q *botapi.CallbackQuery) (err error) {
	if !g.Is("request") {
		return
	}

	msg := g.NewMessageUpdate(
		api.AtUserString(g.Players[0])+" wants to play 🪨 📜 ✂️",
		api.InlineKeyboard([]map[string]string{{"Play!": g.getCmd("accept")}}),
	)

	if _, err = g.Bot.Send(msg); err != nil {
		log.Printf("Error in RequestToss(): %q", err.Error())
		return
	}

	return
}

func (g *RockPaperScissors) AcceptGame(c *api.Context, q *botapi.CallbackQuery) (err error) {
	if !g.Is("request") || (q.Message.Chat.Type != "private" && q.From.ID == g.ID) {
		log.Println("AcceptGame failed")
		return
	}

	g.Mutate("play", q.Message)
	g.Players[1] = q.From

	if err := g.SendRound(c, q); err != nil {
		g.Mutate("request", q.Message)
		g.Players[1] = nil
	}

	return
}

func (g *RockPaperScissors) SendRound(c *api.Context, q *botapi.CallbackQuery) (err error) {
	if !g.Is("play") {
		return
	}

	if g.Round++; g.Round > g.TotalRounds {
		if err = g.SendWinner(c, q); err == nil {
			rpsRunning.Delete(g.ID)
		}

		return
	}

	p1, p2 := api.DisplayName(g.Players[0]), api.DisplayName(g.Players[1])

	mu := botapi.NewInlineKeyboardMarkup(
		[]botapi.InlineKeyboardButton{
			botapi.NewInlineKeyboardButtonData(p1+" "+string(MoveRock), g.getCmd(string(MoveRock))+"/0"),
			botapi.NewInlineKeyboardButtonData(string(MoveRock)+" "+p2, g.getCmd(string(MoveRock))+"/1"),
		},
		[]botapi.InlineKeyboardButton{
			botapi.NewInlineKeyboardButtonData(p1+" "+string(MovePaper), g.getCmd(string(MovePaper))+"/0"),
			botapi.NewInlineKeyboardButtonData(string(MovePaper)+" "+p2, g.getCmd(string(MovePaper))+"/1"),
		},
		[]botapi.InlineKeyboardButton{
			botapi.NewInlineKeyboardButtonData(p1+" "+string(MoveScissors), g.getCmd(string(MoveScissors))+"/0"),
			botapi.NewInlineKeyboardButtonData(string(MoveScissors)+" "+p2, g.getCmd(string(MoveScissors))+"/1"),
		},
	)

	m := g.NewMessageUpdate(g.menuBuilder().String(), &mu)

	if err = api.SendUpdate(g.Bot, m); err != nil {
		g.Round--
	}

	return
}

func (g *RockPaperScissors) DoMove(move Move, i int, q *botapi.CallbackQuery) (done bool, err error) {
	if !g.Is("play") {
		return
	}

	j := g.Round - 1

	if q.From.ID == g.Players[i].ID && g.Moves[j][i] == "" {
		p1, p2 := api.DisplayName(g.Players[0]), api.DisplayName(g.Players[1])
		if i == 0 || g.Moves[j][0] != "" {
			p1 = "✅"
		}
		if i == 1 || g.Moves[j][1] != "" {
			p2 = "✅"
		}

		mu := botapi.NewInlineKeyboardMarkup(
			[]botapi.InlineKeyboardButton{
				botapi.NewInlineKeyboardButtonData(p1+" "+string(MoveRock), g.getCmd(string(MoveRock))+"/0"),
				botapi.NewInlineKeyboardButtonData(string(MoveRock)+" "+p2, g.getCmd(string(MoveRock))+"/1"),
			},
			[]botapi.InlineKeyboardButton{
				botapi.NewInlineKeyboardButtonData(p1+" "+string(MovePaper), g.getCmd(string(MovePaper))+"/0"),
				botapi.NewInlineKeyboardButtonData(string(MovePaper)+" "+p2, g.getCmd(string(MovePaper))+"/1"),
			},
			[]botapi.InlineKeyboardButton{
				botapi.NewInlineKeyboardButtonData(p1+" "+string(MoveScissors), g.getCmd(string(MoveScissors))+"/0"),
				botapi.NewInlineKeyboardButtonData(string(MoveScissors)+" "+p2, g.getCmd(string(MoveScissors))+"/1"),
			},
		)

		m := g.NewMessageUpdate(g.menuBuilder().String(), &mu)
		if err = api.SendUpdate(g.Bot, m); err == nil {
			g.Moves[j][i] = move
		}

	}

	done = g.Moves[j][0] != "" && g.Moves[j][1] != ""
	return
}

func (g *RockPaperScissors) SendWinner(c *api.Context, q *botapi.CallbackQuery) (err error) {
	p1, p2 := 0, 0

	for _, move := range g.Moves {
		if cmp := move[0].Compare(move[1]); cmp > 0 {
			p1 += 1
		} else if cmp < 0 {
			p2 += 1
		}
	}

	text := g.menuBuilder()
	text.WriteString("\n" + fmt.Sprintf("%s %d - %d %s\n", api.AtUserString(g.Players[0]), p1, p2, api.AtUserString(g.Players[1])))

	s := service.NewUserXPService(c.Server.DB)

	winner := "Draw 🥴"
	if p1 > p2 {
		xp := 100 * int64(p1-p2)
		winner = fmt.Sprintf("%s wins! +%d XP", api.AtUserString(g.Players[0]), xp)
		s.UpdateXPs(g.Players[1], service.XPTitleGames, xp)
	} else if p1 < p2 {
		xp := 100 * int64(p2-p1)
		winner = fmt.Sprintf("%s wins! +%d XP", api.AtUserString(g.Players[1]), xp)
		s.UpdateXPs(g.Players[1], service.XPTitleGames, xp)
	}

	text.WriteString(winner)

	m := g.NewMessageUpdate(text.String(), nil)
	err = api.SendUpdate(g.Bot, m)

	return
}

func (a Move) Compare(b Move) int {
	if a == b {
		return 0
	}

	if a == MoveRock && b == MoveScissors ||
		a == MoveScissors && b == MovePaper ||
		a == MovePaper && b == MoveRock {
		return 1
	}

	return -1
}

func (g *RockPaperScissors) getCmd(cmd string) string {
	return fmt.Sprintf("%s/%s/%d", RockPaperScissorsPath, cmd, g.ID)
}

func (g *RockPaperScissors) playerPrefix() string {
	return api.AtString("(P1) "+api.DisplayName(g.Players[0]), g.Players[0].ID) +
		" vs " +
		api.AtString(api.DisplayName(g.Players[1])+" (P2)", g.Players[1].ID)
}

func (g *RockPaperScissors) menuBuilder() *strings.Builder {
	results := map[int]string{
		-1: "🔴 🟢",
		0:  "⚪️ ⚪️",
		1:  "🟢 🔴",
	}

	text := &strings.Builder{}
	text.WriteString(RockPaperScissorsTitle + "\n" + g.playerPrefix() + "\n\n")

	for i := 0; i < g.Round-1; i++ {
		cmp := g.Moves[i][0].Compare(g.Moves[i][1])
		text.WriteString(string(g.Moves[i][0]) + " " + string(g.Moves[i][1]) + " | " + results[cmp] + "\n")
	}

	return text
}

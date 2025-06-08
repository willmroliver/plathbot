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

const (
	ConnectFourTitle = "🟣🟠 Connect 4"
	ConnectFourPath  = Path + "/connect4"
)

type ConnectNode struct {
	Colour     string
	Neighbours [8]*ConnectNode
	ChainLens  [4]int
}

type ConnectFour struct {
	*api.Interaction[string]
	ID      int64
	Board   [7][6]*ConnectNode
	Height  [7]int
	Bot     *botapi.BotAPI
	Players [2]*botapi.User
	Turn    byte
	Mu      sync.Mutex
}

var (
	Colours = [2]string{"🟣", "🟠"}
	Cols    = map[string]int{
		"0": 0,
		"1": 1,
		"2": 2,
		"3": 3,
		"4": 4,
		"5": 5,
		"6": 6,
	}

	c4Running    = sync.Map{}
	neighbourMap = map[int][2]int{
		0: {1, 0},
		1: {1, -1},
		2: {0, -1},
		3: {-1, -1},
		4: {-1, 0},
		5: {-1, 1},
		6: {0, 1},
		7: {1, 1},
	}
)

func ConnectFourQuery(c *api.Context, query *botapi.CallbackQuery, cmd *api.CallbackCmd) {
	c4Running.Range(func(key any, value any) bool {
		game := value.(*ConnectFour)
		if game.Age() > time.Minute*20 {
			c4Running.Delete(key)
		}

		return true
	})

	action := cmd.Get()

	if action == "" {
		_, exists := c4Running.Load(query.From.ID)
		if !exists {
			game := NewConnectFour(c, 3)

			if game.RequestGame(query) == nil {
				c4Running.Store(game.ID, game)
			}
		}

		return
	}

	id, err := strconv.ParseInt(cmd.Next().Get(), 10, 64)
	if err != nil {
		log.Printf("Invalid game ID: %q", err.Error())
		return
	}

	val, exists := c4Running.Load(id)
	if !exists {
		log.Printf("Game %d does not exist", id)
		return
	}

	game := val.(*ConnectFour)
	turn := game.Turn

	game.Mu.Lock()
	defer game.Mu.Unlock()

	if turn != game.Turn {
		return
	}

	if action == "accept" {
		if game.AcceptGame(c, query) != nil {
			c4Running.Delete(game.ID)
		}

		return
	}

	if game.Players[turn].ID != c.User.ID {
		return
	}

	col, ok := Cols[action]
	if !ok {
		return
	}

	done, err := game.DoMove(col, c, query)

	if done {
		c4Running.Delete(query.From.ID)
	}
}

func NewConnectFour(c *api.Context, rounds int) *ConnectFour {
	return &ConnectFour{
		Interaction: api.NewInteraction[string](c.Message, "request"),
		ID:          c.User.ID,
		Bot:         c.Bot,
		Players:     [2]*botapi.User{c.User, nil},
	}
}

func (g *ConnectFour) RequestGame(q *botapi.CallbackQuery) (err error) {
	if !g.Is("request") {
		return
	}

	msg := g.NewMessageUpdate(
		api.AtUserString(g.Players[0])+" wants to play 🟣🟠🟣🟠",
		api.InlineKeyboard([]map[string]string{{"Play!": g.getCmd("accept")}}),
	)

	if _, err = g.Bot.Send(msg); err != nil {
		log.Printf("Error in RequestToss(): %q", err.Error())
		return
	}

	return
}

func (g *ConnectFour) AcceptGame(c *api.Context, q *botapi.CallbackQuery) (err error) {
	if !g.Is("request") {
		log.Println("AcceptGame failed")
		return
	}

	g.Mutate("play", q.Message)
	g.Players[1] = q.From

	if err := g.SendBoard(c, q); err != nil {
		g.Mutate("request", q.Message)
		g.Players[1] = nil
	}

	return
}

func (g *ConnectFour) SendBoard(c *api.Context, q *botapi.CallbackQuery) (err error) {
	if !g.Is("play") {
		return
	}

	m := g.NewMessageUpdate(
		g.menuBuilder().String(),
		g.movesKeyboard(),
	)

	err = api.SendUpdate(g.Bot, m)
	return
}

func (g *ConnectFour) DoMove(col int, c *api.Context, q *botapi.CallbackQuery) (done bool, err error) {
	if !g.Is("play") {
		return
	}

	colour := Colours[g.Turn]
	n := &ConnectNode{
		Colour:    colour,
		ChainLens: [4]int{1, 1, 1, 1},
	}

	row := g.Height[col]
	g.Board[col][row] = n

	for i, coord := range neighbourMap {
		if m := g.getNode(col+coord[0], row+coord[1]); m != nil && m.Colour == n.Colour {
			n.Neighbours[i] = m
			m.Neighbours[(i+4)%8] = n

			chainLen := n.ChainLens[i%4] + m.ChainLens[i%4]
			if chainLen > 3 {
				for o := m; o != nil && o.Colour == colour; o = o.Neighbours[i] {
					o.Colour = "🟢"
				}
				for o := n; o != nil && o.Colour == colour; o = o.Neighbours[(i+4)%8] {
					o.Colour = "🟢"
				}

				g.SendWinner(c, q, int(g.Turn))
				done = true
				return
			}

			for o := m; o != nil && o.Colour == colour; o = o.Neighbours[i] {
				o.ChainLens[i%4] = chainLen
			}
			for o := n; o != nil && o.Colour == colour; o = o.Neighbours[(i+4)%8] {
				o.ChainLens[i%4] = chainLen
			}
		}
	}

	g.Height[col]++
	g.Turn = 1 - g.Turn

	m := g.NewMessageUpdate(
		g.menuBuilder().String(),
		g.movesKeyboard(),
	)

	moveMux.Lock()
	defer func() {
		time.Sleep(time.Millisecond * 400)
		moveMux.Unlock()
	}()

	if err = api.SendUpdate(g.Bot, m); err != nil {
		g.Height[col]--
		g.Turn = 1 - g.Turn
		return
	}

	for _, n := range g.Height {
		if n < 6 {
			return
		}
	}

	g.SendWinner(c, q, -1)
	return
}

func (g *ConnectFour) SendWinner(c *api.Context, q *botapi.CallbackQuery, turn int) (err error) {
	text := g.menuBuilder()

	s := service.NewUserXPService(c.Server.DB)

	winner := "Draw 🥴"

	if turn != -1 {
		xp := int64(100)
		winner = fmt.Sprintf("%s wins! %s +%d XP", api.AtUserString(g.Players[turn]), Colours[turn], xp)
		s.UpdateXPs(g.Players[turn], service.XPTitleGames, xp)
	}

	text.WriteString(winner)

	m := g.NewMessageUpdate(text.String(), nil)
	err = api.SendUpdate(g.Bot, m)

	return
}

func (g *ConnectFour) getNode(x, y int) *ConnectNode {
	if x < 0 || x > 6 || y < 0 || y > 5 {
		return nil
	}

	return g.Board[x][y]
}

func (g *ConnectFour) getCmd(cmd string) string {
	return fmt.Sprintf("%s/%s/%d", ConnectFourPath, cmd, g.ID)
}

func (g *ConnectFour) playerPrefix() string {
	return api.AtString("(P1) "+api.DisplayName(g.Players[0]), g.Players[0].ID) +
		" vs " +
		api.AtString(api.DisplayName(g.Players[1])+" (P2)", g.Players[1].ID)
}

func (g *ConnectFour) menuBuilder() *strings.Builder {
	text := &strings.Builder{}
	text.WriteString(ConnectFourTitle + "\n" + g.playerPrefix() + "\n\n")

	for i := range 6 {
		for j := range 7 {
			colour := "⚪️"
			if n := g.getNode(j, 5-i); n != nil {
				colour = n.Colour
			}

			text.WriteString("  " + colour + "  ")
		}
		text.WriteString("\n\n")
	}

	return text
}

func (g *ConnectFour) movesKeyboard() *botapi.InlineKeyboardMarkup {
	pl := api.DisplayName(g.Players[g.Turn])
	moves := [7]botapi.InlineKeyboardButton{}

	for i := range moves {
		if g.Height[i] == 6 {
			moves[i] = botapi.NewInlineKeyboardButtonData("✅", g.getCmd("ignore"))
		} else {
			moves[i] = botapi.NewInlineKeyboardButtonData("⬆️", g.getCmd(fmt.Sprintf("%d", i)))
		}
	}

	mu := botapi.NewInlineKeyboardMarkup(
		moves[:],
		[]botapi.InlineKeyboardButton{
			botapi.NewInlineKeyboardButtonData(pl+" "+Colours[g.Turn], g.getCmd("ignore")),
		},
	)

	return &mu
}

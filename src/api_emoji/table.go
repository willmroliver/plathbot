package emoji

import (
	"fmt"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
)

const (
	TableTitle = "📊 Rankings"
	TablePath  = Path + "/table"
)

func TableAPI() *api.CallbackAPI {
	return api.NewCallbackAPI(
		TableTitle,
		TablePath,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"all":   getAll,
				"month": getMonthly,
				"week":  getWeekly,
			},
			PublicOptions: []map[string]string{
				{"⏳ All-Time": "all"},
				{"📆 Monthly": "month"},
				{"📰 This Week": "week"},
				api.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

func getAll(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r := repo.NewReactCountRepo(c.Server.DB)
	sendTable(c, "⏳ All-Time Leaderboard", r.TopCounts())
}

func getMonthly(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r := repo.NewReactCountRepo(c.Server.DB)
	sendTable(c, "📆 Monthly Leaderboard", r.TopMonthly())
}

func getWeekly(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r := repo.NewReactCountRepo(c.Server.DB)
	sendTable(c, "📰 Weekly Leaderboard", r.TopWeekly())
}

func sendTable(c *api.Context, title string, data []*model.ReactCount) {
	r := repo.NewReactRepo(c.Server.DB)

	text := &strings.Builder{}
	text.WriteString(title + "\n\n")

	for _, count := range data {
		if react := r.Get(count.Emoji); react != nil {
			text.WriteString(fmt.Sprintf(
				"%s %d\t %s - %s\n",
				count.Emoji,
				count.Count,
				api.AtString(count.User.FirstName, count.User.ID),
				react.Title,
			))
		}
	}

	msg := botapi.NewEditMessageTextAndMarkup(
		c.Chat.ID,
		c.Message.MessageID,
		text.String(),
		*api.InlineKeyboard([]map[string]string{api.KeyboardNavRow(TablePath)}),
	)
	msg.ParseMode = "Markdown"

	api.SendUpdate(c.Bot, &msg)
}

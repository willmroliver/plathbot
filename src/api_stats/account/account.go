//go:build stats && account
// +build stats,account

package account

import (
	"strconv"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	account "github.com/willmroliver/plathbot/src/api_account"
	stats "github.com/willmroliver/plathbot/src/api_stats"
	"github.com/willmroliver/plathbot/src/repo"
)

func init() {
	account.Extensions.ExtendAPI(stats.Title, stats.Path, API)
}

func API(c *api.Context, query *botapi.CallbackQuery, cmd *api.CallbackCmd) {
	titles := repo.NewUserXPRepo(c.Server.DB).Titles()
	user := c.GetUser()

	data := make([]string, 4*(len(titles)+2))
	data[0] = stats.Title
	data[1] = "Week"
	data[2] = "Month"
	data[3] = "All"

	i := 8

	for _, title := range titles {
		if j := strings.LastIndex(title, " "); i != -1 {
			data[i] = title[:j]
		} else {
			data[i] = title
		}

		xp := user.UserXPMap[title]

		for range 3 {
			data[i+1] = strconv.FormatInt(xp.WeekXP, 10)
			data[i+2] = strconv.FormatInt(xp.MonthXP, 10)
			data[i+3] = strconv.FormatInt(xp.XP, 10)
		}

		i += 4
	}

	msg := botapi.NewEditMessageText(
		c.Chat.ID,
		c.Message.MessageID,
		api.MarkdownV2Cols(data, 4),
	)

	msg.ReplyMarkup = api.InlineKeyboard([]map[string]string{api.KeyboardNavRow(account.Path)})
	msg.ParseMode = botapi.ModeMarkdownV2

	api.SendUpdate(c.Bot, &msg)
}

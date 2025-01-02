package stats

import (
	"fmt"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/service"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	XpTitle = "📈 XP"
	XpPath  = Path + "/xp"
)

func UserXPAPI() *api.CallbackAPI {
	return api.NewCallbackAPI(
		XpTitle,
		XpPath,
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
				util.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

func getAll(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r := repo.NewUserXPRepo(c.Server.DB)
	sendTable(
		c,
		"⏳ All-Time",
		r.TopXPs(service.XPTitleEngage, "xp DESC", 0, 5),
		func(xp *model.UserXP) int64 {
			return xp.XP
		},
	)
}

func getMonthly(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r := repo.NewUserXPRepo(c.Server.DB)
	sendTable(
		c, "📆 Monthly",
		r.TopXPs(service.XPTitleEngage, "month_xp DESC", 0, 5),
		func(xp *model.UserXP) int64 {
			return xp.MonthXP
		},
	)
}

func getWeekly(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r := repo.NewUserXPRepo(c.Server.DB)
	sendTable(
		c,
		"📰 Weekly",
		r.TopXPs(service.XPTitleEngage, "week_xp DESC", 0, 5),
		func(xp *model.UserXP) int64 {
			return xp.WeekXP
		},
	)
}

func sendTable(c *api.Context, title string, data []*model.UserXP, get func(*model.UserXP) int64) {
	if len(data) == 0 {
		return
	}

	text := &strings.Builder{}
	text.WriteString(title + " - " + data[0].Title + "\n\n")

	for i, xp := range data {
		if i == 0 {
			text.WriteString(fmt.Sprintf(
				"👑. %s - %d\n",
				util.AtString(xp.User.FirstName, xp.User.ID),
				get(xp),
			))
			continue
		}

		text.WriteString(fmt.Sprintf(
			"%d. %s - %d\n",
			i+1,
			util.AtString(xp.User.FirstName, xp.User.ID),
			get(xp),
		))
	}

	msg := botapi.NewEditMessageTextAndMarkup(
		c.Chat.ID,
		c.Message.MessageID,
		text.String(),
		*util.InlineKeyboard([]map[string]string{util.KeyboardNavRow(XpPath)}),
	)
	msg.ParseMode = "Markdown"

	util.SendUpdate(c.Bot, &msg)
}

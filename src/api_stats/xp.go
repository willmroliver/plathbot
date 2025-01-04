package stats

import (
	"fmt"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
)

const (
	XpTitle = "📈 XP"
	XpPath  = Path
)

func UserXPAPI(title string) *api.CallbackAPI {
	all, month, week := "all", "week", "month"

	return api.NewCallbackAPI(
		title,
		XpPath+"/"+title,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				all:   XPTitle(title).getAll,
				month: XPTitle(title).getMonthly,
				week:  XPTitle(title).getWeekly,
			},
			PublicOptions: []map[string]string{
				{"⏳ All-Time": all},
				{"📆 Monthly": month},
				{"📰 This Week": week},
				api.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

type XPTitle string

func (t XPTitle) getAll(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r := repo.NewUserXPRepo(c.Server.DB)
	t.sendTable(
		c,
		"⏳ All-Time",
		r.TopXPs(string(t), "xp DESC", 0, 5),
		func(xp *model.UserXP) int64 {
			return xp.XP
		},
	)
}

func (t XPTitle) getMonthly(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r := repo.NewUserXPRepo(c.Server.DB)
	t.sendTable(
		c, "📆 Monthly",
		r.TopXPs(string(t), "month_xp DESC", 0, 5),
		func(xp *model.UserXP) int64 {
			return xp.MonthXP
		},
	)
}

func (t XPTitle) getWeekly(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r := repo.NewUserXPRepo(c.Server.DB)
	t.sendTable(
		c,
		"📰 Weekly",
		r.TopXPs(string(t), "week_xp DESC", 0, 5),
		func(xp *model.UserXP) int64 {
			return xp.WeekXP
		},
	)
}

func (t XPTitle) sendTable(c *api.Context, title string, data []*model.UserXP, get func(*model.UserXP) int64) {
	if len(data) == 0 {
		return
	}

	text := &strings.Builder{}
	text.WriteString(title + " - " + data[0].Title + "\n\n")

	for i, xp := range data {
		if i == 0 {
			text.WriteString(fmt.Sprintf(
				"👑 %s - %d\n",
				api.AtString(xp.User.FirstName, xp.User.ID),
				get(xp),
			))
			continue
		}

		text.WriteString(fmt.Sprintf(
			"%d. %s - %d\n",
			i+1,
			api.AtString(xp.User.FirstName, xp.User.ID),
			get(xp),
		))
	}

	msg := botapi.NewEditMessageTextAndMarkup(
		c.Chat.ID,
		c.Message.MessageID,
		text.String(),
		*api.InlineKeyboard([]map[string]string{api.KeyboardNavRow(XpPath + "/" + string(t))}),
	)
	msg.ParseMode = "Markdown"

	api.SendUpdate(c.Bot, &msg)
}

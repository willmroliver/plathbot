package reddit

import (
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
)

const (
	Title = "🤖 Reddit"
	Path  = "reddit"
)

var (
	adminAPI = AdminAPI()
)

func API() *api.CallbackAPI {
	admin, view := "admin", "raid"

	return api.NewCallbackAPI(
		Title,
		Path,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				admin: adminAPI.Select,
				view:  allPosts,
			},
			PublicCooldown: time.Second * 3,
			PublicOptions: []map[string]string{
				{AdminTitle: admin},
				{"🤑 Raid!": view},
				api.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

func allPosts(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) {
	r, posts := repo.NewRepo(c.Server.DB), []*model.RedditPost{}

	c.Server.DB.Where("expires_at < ?", time.Now()).Delete(&model.RedditPost{})

	if err := r.All(&posts); err != nil || len(posts) == 0 {
		return
	}

	text := "‼‼‼ *Raid Links* ‼‼‼\nTop shillers _will be rewarded_ 🤑"
	kb := make([]map[string]string, len(posts))

	for i, p := range posts {
		kb[i] = map[string]string{p.Title: api.KeyboardLink(p.URL)}
	}

	m := botapi.NewEditMessageTextAndMarkup(c.Chat.ID, c.Message.MessageID, text, *api.InlineKeyboard(kb))
	m.ParseMode = botapi.ModeMarkdown

	api.SendConfig(c.Bot, m)
}

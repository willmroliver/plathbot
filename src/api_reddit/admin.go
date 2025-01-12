package reddit

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

const (
	AdminTitle = "🔐 Manage"
	AdminPath  = Path + "/admin"
)

var open = sync.Map{}

func AdminAPI() *api.CallbackAPI {
	add, view, remove := "add", "view", "remove"

	return api.NewCallbackAPI(
		AdminTitle,
		AdminPath,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				add: func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if a := OpenAdmin(c, cq, cc); a != nil {
						a.Update(c, cq)
					}
				},
				view: func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if a := OpenAdmin(c, cq, cc); a != nil {
						a.View(c, cq)
					}
				},
				remove: func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if a := OpenAdmin(c, cq, cc); a != nil {
						a.Remove(c, cq, cc)
					}
				},
			},
			PublicOptions: []map[string]string{
				{"✏️ Add Post": add},
				{"👀 View Active": view},
				{"🔚 Stop Tracking": remove},
				api.KeyboardNavRow(".."),
			},
			PublicOnly: true,
		},
	)
}

type Admin struct {
	*api.Interaction[string]
	user *botapi.User
	repo *repo.Repo
}

func NewAdmin(db *gorm.DB, q *botapi.CallbackQuery) *Admin {
	return &Admin{
		api.NewInteraction(q.Message, ""),
		q.From,
		repo.NewRepo(db),
	}
}

func OpenAdmin(c *api.Context, q *botapi.CallbackQuery, cc *api.CallbackCmd) (admin *Admin) {
	if u, ok := cc.Tags["user"]; ok {
		if id, _ := strconv.ParseInt(u, 10, 64); id != q.From.ID {
			return
		}
	}

	if !c.IsAdmin() {
		return
	}

	open.Range(func(key any, value any) bool {
		if value.(*api.Interaction[any]).Age() > time.Minute*5 {
			open.Delete(key)
		}

		return true
	})

	if data, exists := open.Load(q.From.ID); exists {
		admin = data.(*Admin)
	} else {
		admin = NewAdmin(c.Server.DB, q)
	}

	return
}

func (a *Admin) View(c *api.Context, query *botapi.CallbackQuery) {
	mu := api.InlineKeyboard([]map[string]string{
		api.KeyboardNavRow(AdminPath),
	}, fmt.Sprintf("user=%d", a.user.ID))

	posts := []*model.RedditPost{}
	if err := a.repo.All(&posts); err != nil {
		api.SendUpdate(c.Bot, a.NewMessageUpdate("Error fetching posts.", mu))
		return
	}

	if len(posts) == 0 {
		api.SendUpdate(c.Bot, a.NewMessageUpdate("No posts being tracked.", mu))
		return
	}

	now := time.Now()
	text := "👀 Active Posts\n\n"

	for _, post := range posts {
		text += fmt.Sprintf("%s - %s\n", post.Title, post.ExpiresAt.Sub(now).String())
	}

	api.SendUpdate(c.Bot, a.NewMessageUpdate(text, mu))
}

func (a *Admin) Update(c *api.Context, query *botapi.CallbackQuery) {
	api.SendUpdate(c.Bot, a.NewMessageUpdate(`
Okay, send the post ID you'd like to start tracking. ID can be found in the URL, E.g:

/r/SolanaMemeCoins/comments/1hetkr8/plath_holding_strong/ -> '1hetkr8'
	`, nil))

	hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, data any) (done bool) {
		ad, post := data.(*Admin), GetPost(m.Text)

		if post == nil {
			api.SendBasic(c.Bot, c.Chat.ID, "Invalid post ID.")
		}

		r := model.NewRedditPost(post)

		api.SendUpdate(c.Bot, ad.NewMessageUpdate(`
How long do you want to track this post for? E.g:

'24h'
'1h30m'
'3h 15m 30 s'
		`, nil))

		hook, ch := api.GetDurationHook(c.Chat.ID, time.Minute*5)
		s.RegisterUserHook(c.User.ID, hook)

		text := "✅ Post added"

		select {
		case dur := <-ch:
			r.ExpiresAt = time.Now().Add(dur)

			if ad.repo.Save(r) != nil {
				text = "Error saving post."
			}
		case <-time.After(time.Minute * 5):
			text = "Post tracking cancelled."
		}

		mu := api.InlineKeyboard([]map[string]string{
			api.KeyboardNavRow(AdminPath),
		}, fmt.Sprintf("user=%d", ad.user.ID))

		api.SendConfig(c.Bot, a.NewMessage(text, mu))

		done = true
		return
	}, a, time.Minute*5)

	c.Server.RegisterUserHook(c.User.ID, hook)
}

func (a *Admin) Remove(c *api.Context, query *botapi.CallbackQuery, cc *api.CallbackCmd) {
	mu := api.InlineKeyboard([]map[string]string{
		api.KeyboardNavRow(AdminPath),
	}, fmt.Sprintf("user=%d", c.User.ID))

	if postID := cc.Get(); postID != "" {
		if err := a.repo.Delete(&model.RedditPost{PostID: postID}); err != nil {
			api.SendUpdate(c.Bot, a.NewMessageUpdate("Error removing post.", mu))
		} else {
			api.SendUpdate(c.Bot, a.NewMessageUpdate("✅ Post removed", mu))
		}

		return
	}

	var opts []map[string]string

	posts := []*model.RedditPost{}
	if err := a.repo.All(&posts); err == nil {
		opts = make([]map[string]string, len(posts)+1)
		for i, post := range posts {
			opts[i] = map[string]string{post.Title: AdminPath + "/remove/" + post.PostID}
		}
		opts[len(posts)] = api.KeyboardNavRow(AdminPath)
	}

	mu = api.InlineKeyboard(opts, fmt.Sprintf("user=%d", c.User.ID))
	api.SendUpdate(c.Bot, a.NewMessageUpdate("Select a post to remove", mu))
}

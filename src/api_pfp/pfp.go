package pfp

import (
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/repo"
)

func API(c *api.Context, m *botapi.Message, args ...string) {
	if len(args) == 0 {
		random(c, m, args...)
		return
	}

	actions := map[string]api.CommandAction{
		"":       random,
		"add":    add,
		"list":   list,
		"get":    get,
		"delete": delete,
	}

	if action, ok := actions[args[0]]; ok {
		action(c, m, args...)
	}
}

func random(c *api.Context, m *botapi.Message, args ...string) {
	r := repo.NewFileRepo(c.Server.DB)
	if files := r.List("/pfp/", "RANDOM()", 0, 1, "Photo"); len(files) != 0 {
		m := botapi.NewPhoto(c.Chat.ID, botapi.FileID(files[0].FileID))
		api.SendConfig(c.Bot, m)
	}
}

func add(c *api.Context, m *botapi.Message, args ...string) {
	api.SendBasic(c.Bot, c.Chat.ID, "Send a name & image you'd like to add to /pfp")

	var title string
	var photo *botapi.PhotoSize

	hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, a any) (done bool) {
		title = m.Text

		if len(m.Photo) != 0 {
			photo = &m.Photo[0]
			title = m.Caption
		}

		if done = title != "" && photo != nil; done {
			r := repo.NewFileRepo(c.Server.DB)
			if r.Save(photo, "/pfp/"+title) != nil {
				api.SendBasic(c.Bot, c.Chat.ID, "Image added to /pfp.")
			} else {
				api.SendBasic(c.Bot, c.Chat.ID, "Oops, something went wrong.")
			}
		} else if title != "" {
			api.SendBasic(c.Bot, c.Chat.ID, "Perfect, now send an image.")
		} else {
			api.SendBasic(c.Bot, c.Chat.ID, "Great, now send a name for the image.")
		}

		return
	}, c.User.ID, time.Minute*5)

	c.Server.RegisterUserHook(c.User.ID, hook)
}

func list(c *api.Context, m *botapi.Message, args ...string)   {}
func get(c *api.Context, m *botapi.Message, args ...string)    {}
func delete(c *api.Context, m *botapi.Message, args ...string) {}

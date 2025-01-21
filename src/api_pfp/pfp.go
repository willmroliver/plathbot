package pfp

import (
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/model"
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
		action(c, m, args[1:]...)
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

func list(c *api.Context, m *botapi.Message, args ...string) {
	r := repo.NewFileRepo(c.Server.DB)

	files := r.List("/pfp", "", 0, 1000, "Photo")
	if files == nil {
		return
	} else if len(files) == 0 {
		api.SendBasic(c.Bot, c.Chat.ID, "No PFPs found :(")
		return
	}

	msg := botapi.NewMessage(c.Chat.ID, "🎨 PFPs 📸")

	var mu [][]botapi.InlineKeyboardButton

	if c.IsAdmin() && m.Chat.Type == "private" && false {
		mu = make([][]botapi.InlineKeyboardButton, len(files))

		for i, f := range files {
			mu[i] = []botapi.InlineKeyboardButton{
				botapi.NewInlineKeyboardButtonData(f.Name+" 👀", "cmd|/pfp get "+f.FileUniqueID),
				botapi.NewInlineKeyboardButtonData("🗑️", "cmd|/pfp delete "+f.FileUniqueID),
			}
		}
	} else {
		mu = make([][]botapi.InlineKeyboardButton, (len(files)+2)/3)
		row := make([]botapi.InlineKeyboardButton, 0, 3)

		for i, f := range files {
			row = append(row, botapi.NewInlineKeyboardButtonData(f.Name[5:]+" 👀", "cmd|/pfp get "+f.FileUniqueID))

			if i%3 == 2 || i+1 == len(files) {
				mu[i/3] = row
				row = make([]botapi.InlineKeyboardButton, 0, 3)
			}
		}
	}

	msg.ReplyMarkup = botapi.NewInlineKeyboardMarkup(mu...)
	api.SendConfig(c.Bot, msg)
}

func get(c *api.Context, m *botapi.Message, args ...string) {
	if len(args) == 0 {
		return
	}

	file := &model.File{}

	if repo.NewFileRepo(c.Server.DB).GetBy(file, "file_unique_id", args[0]) == nil {
		m := botapi.NewPhoto(c.Chat.ID, botapi.FileID(file.FileID))
		api.SendConfig(c.Bot, m)
	}
}

func delete(c *api.Context, m *botapi.Message, args ...string) {
	if !c.IsAdmin() || len(args) == 0 {
		return
	}

	repo.NewFileRepo(c.Server.DB).DeleteBy(&model.File{}, "file_unique_id", args[0])

	api.SendBasic(c.Bot, c.Chat.ID, "✅ File deleted")
}

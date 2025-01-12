package account

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	goreddit "github.com/vartanbeno/go-reddit/v2/reddit"
	"github.com/willmroliver/plathbot/src/api"
	reddit "github.com/willmroliver/plathbot/src/api_reddit"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"gorm.io/gorm"
)

const (
	RedditTitle = "🤖 Reddit"
	RedditPath  = Path + "/reddit"
)

var redditsOpen = sync.Map{}

func RedditAPI() *api.CallbackAPI {
	return api.NewCallbackAPI(
		RedditTitle,
		RedditPath,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"view": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if r := OpenReddit(c.Server.DB, cq); r != nil {
						r.View(c, cq)
					}
				},
				"update": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if r := OpenReddit(c.Server.DB, cq); r != nil {
						r.Update(c, cq)
					}
				},
				"remove": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if r := OpenReddit(c.Server.DB, cq); r != nil {
						r.Remove(c, cq)
					}
				},
			},
			PrivateOptions: []map[string]string{
				{"🔗 Link Account": "update"},
				{"👀 View": "view", "😶‍🌫️ Unlink": "remove"},
				api.KeyboardNavRow(".."),
			},
			PrivateOnly: true,
		},
	)
}

type Reddit struct {
	*api.Interaction[string]
	repo *repo.UserRepo
	user *botapi.User
}

func NewReddit(db *gorm.DB, query *botapi.CallbackQuery) *Reddit {
	return &Reddit{
		Interaction: api.NewInteraction[string](query.Message, ""),
		repo:        repo.NewUserRepo(db),
		user:        query.From,
	}
}

func OpenReddit(db *gorm.DB, query *botapi.CallbackQuery) (r *Reddit) {
	redditsOpen.Range(func(key any, value any) bool {
		if value.(*api.Interaction[any]).Age() > time.Minute*5 {
			redditsOpen.Delete(key)
		}

		return true
	})

	if data, exists := redditsOpen.Load(query.From.ID); exists {
		r = data.(*Reddit)
	} else {
		r = NewReddit(db, query)
	}

	return
}

func (r *Reddit) User() *model.User {
	return r.repo.Get(r.user)
}

func (r *Reddit) View(c *api.Context, query *botapi.CallbackQuery) {
	text := r.User().RedditUsername
	if text == "" {
		text = "Not linked"
	}

	api.SendBasic(c.Bot, query.Message.Chat.ID, text)
}

func (r *Reddit) Update(c *api.Context, query *botapi.CallbackQuery) {
	r.Mutate("update", query.Message)

	api.SendConfig(c.Bot, r.NewMessage("Okay! Send me the username of the reddit account you'd like to link.", nil))

	hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, data any) (done bool) {
		done = true
		re := data.(*Reddit)

		if !r.Is("update") {
			return
		}

		// Create a new link post
		post := reddit.PostBasic(
			"🔗 Account Link Request",
			`
Please comment your verification token on this post to complete your account link.

[Disclaimer] This post is part of a Telegram-Reddit verification system.
It will be hidden & locked after verification, and deleted at a later time.
			`,
		)

		if post == nil {
			api.SendConfig(s.Bot, re.NewMessage("Something went wrong. Please try again", nil))
			return
		}

		// Generate a verification token
		bytes := make([]byte, 15)
		if _, err := rand.Read(bytes); err != nil {
			log.Printf("Error generating reddit link token: %q", bytes)
			return
		}

		token := base64.StdEncoding.EncodeToString(bytes)

		// Send the verification token & post
		api.SendConfig(s.Bot, re.NewMessage(token, nil))
		api.SendConfig(s.Bot, re.NewMessage(`
🔗 Account Link Request

1️⃣ Hit the "Verify" button below
2️⃣ Comment the verification token on the post
3️⃣ Come back to this chat to see confirmation
		`,
			api.InlineKeyboard([]map[string]string{{"Verify": api.KeyboardLink(post.URL)}}),
		))

		// Poll the post for the verification token
		type Payload struct {
			Username, Token string
		}

		comments := reddit.PollComments(post.ID, time.Second*4, time.Minute*5, func(c *goreddit.PostAndComments, payload any) bool {
			for _, comment := range c.Comments {
				if comment.Body == token && comment.Author == payload.(*Payload).Username {
					c.Comments = []*goreddit.Comment{comment}
					return true
				}
			}

			return false
		}, &Payload{Username: m.Text, Token: string(token)})

		// If the verification token was found, link the account
		if comments != nil {
			user := c.GetUser()
			user.RedditUsername = comments[0].Author

			if err := c.UserRepo.Save(user); err != nil {
				api.SendConfig(s.Bot, re.NewMessage("Something went wrong. Please try again", nil))
				return
			}

			api.SendConfig(s.Bot, re.NewMessage(
				fmt.Sprintf("✅ Linked to %q", user.RedditUsername),
				api.InlineKeyboard([]map[string]string{{Title: Path}}, fmt.Sprintf("user=%d", re.user.ID)),
			))
		} else {
			api.SendConfig(s.Bot, re.NewMessage("Verification failed. Please try again.", nil))
		}

		// Either way, lock & hide the post
		reddit.LockPost(post.FullID)
		reddit.HidePost(post.FullID)

		return
	}, r, time.Minute*5)

	c.Server.RegisterChatHook(query.Message.Chat.ID, hook)
}

func (r *Reddit) Remove(c *api.Context, query *botapi.CallbackQuery) {
	user := c.GetUser()
	user.RedditUsername = ""

	if err := c.UserRepo.Save(user); err != nil {
		api.SendBasic(c.Server.Bot, c.Chat.ID, "Unexpected error unlinking account.")
		return
	}

	api.SendUpdate(c.Bot, r.NewMessageUpdate(
		"✅ Deleted",
		api.InlineKeyboard([]map[string]string{{Title: Path}}, fmt.Sprintf("user=%d", r.user.ID)),
	))
}

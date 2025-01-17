package account

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
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

var (
	redditsOpen, confirmOpen = sync.Map{}, sync.Map{}
)

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
				"confirm": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if ch, ok := confirmOpen.LoadAndDelete(c.User.ID); ok {
						ch.(chan struct{}) <- struct{}{}
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

1️⃣ Hit the *Verify* button below

2️⃣ Send the verification token.

3️⃣ Come back here and hit *Confirm* to verify.
			`,
			api.InlineKeyboard([]map[string]string{{
				"Verify": api.KeyboardLink(fmt.Sprintf(
					"https://www.reddit.com/message/compose/?to=%s&subject=Verify&message=%s",
					os.Getenv("GO_REDDIT_CLIENT_USERNAME"),
					token,
				)),
			}, {
				"Confirm": RedditPath + "/confirm",
			}}),
		))

		ch := make(chan struct{}, 1)
		confirmOpen.Store(c.User.ID, ch)

		select {
		case <-ch:
			break
		case <-time.After(time.Minute * 5):
			confirmOpen.Delete(c.User.ID)
			return
		}

		// Poll the post for the verification token
		type Payload struct {
			Username, Token string
			Success         bool
		}

		payload := &Payload{Username: m.Text, Token: string(token)}

		reddit.PollInbox(time.Second, 0, func(ms []*goreddit.Message, messages []*goreddit.Message, p any) bool {
			payload = p.(*Payload)

			for _, m := range append(ms, messages...) {
				if m.Text == payload.Token && m.Author == payload.Username {
					payload.Success = true
					return true
				}
			}

			return false
		}, payload)

		// If the verification token was found, link the account
		if payload.Success {
			user := c.GetUser()
			user.RedditUsername = payload.Username

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

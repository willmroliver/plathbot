package account

import (
	"fmt"
	"sync"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"gorm.io/gorm"
)

const (
	WalletTitle = "💳 Wallet"
	WalletPath  = Path + "/wallet"
)

var walletsOpen = sync.Map{}

func WalletAPI() *api.CallbackAPI {
	return api.NewCallbackAPI(
		WalletTitle,
		WalletPath,
		&api.CallbackConfig{
			Actions: map[string]api.CallbackAction{
				"view": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if w := OpenWallet(c.Server.DB, cq); w != nil {
						w.View(c, cq)
					}
				},
				"update": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if w := OpenWallet(c.Server.DB, cq); w != nil {
						w.Update(c, cq)
					}
				},
				"remove": func(c *api.Context, cq *botapi.CallbackQuery, cc *api.CallbackCmd) {
					if w := OpenWallet(c.Server.DB, cq); w != nil {
						w.Remove(c, cq)
					}
				},
			},
			PrivateOptions: []map[string]string{
				{"✏️ Update": "update"},
				{"👀 View": "view", "🗑️ Remove": "remove"},
				api.KeyboardNavRow(".."),
			},
			PrivateOnly: true,
		},
	)
}

type Wallet struct {
	*api.Interaction[string]
	repo *repo.UserRepo
	user *botapi.User
}

func NewWallet(db *gorm.DB, query *botapi.CallbackQuery) *Wallet {
	return &Wallet{
		Interaction: api.NewInteraction[string](query.Message, ""),
		repo:        repo.NewUserRepo(db),
		user:        query.From,
	}
}

func OpenWallet(db *gorm.DB, query *botapi.CallbackQuery) (wallet *Wallet) {
	walletsOpen.Range(func(key any, value any) bool {
		if value.(*api.Interaction[any]).Age() > time.Minute*5 {
			walletsOpen.Delete(key)
		}

		return true
	})

	if data, exists := walletsOpen.Load(query.From.ID); exists {
		wallet = data.(*Wallet)
	} else {
		wallet = NewWallet(db, query)
	}

	return
}

func (w *Wallet) User() *model.User {
	return w.repo.Get(w.user)
}

func (w *Wallet) View(c *api.Context, query *botapi.CallbackQuery) {
	api.SendBasic(c.Bot, query.Message.Chat.ID, w.User().PublicWallet)
}

func (w *Wallet) Update(c *api.Context, query *botapi.CallbackQuery) {
	w.Mutate("update", query.Message)

	api.SendConfig(c.Bot, w.NewMessage("Okay! Send me a public wallet address to associate to your account.", nil))

	hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, data any) (done bool) {
		done = true
		wa := data.(*Wallet)

		if !w.Is("update") {
			return
		}

		if err := wa.repo.UpdateWallet(wa.user, m.Text); err != nil {
			api.SendConfig(s.Bot, wa.NewMessage("Something went wrong updating your wallet details", nil))
		}

		api.SendConfig(s.Bot, wa.NewMessage(
			"✅ Saved",
			api.InlineKeyboard([]map[string]string{{Title: Path}}, fmt.Sprintf("user=%d", w.user.ID)),
		))

		return
	}, w, time.Minute*5)

	c.Server.RegisterChatHook(query.Message.Chat.ID, hook)
}

func (w *Wallet) Remove(c *api.Context, query *botapi.CallbackQuery) {
	if err := w.repo.UpdateWallet(w.user, ""); err != nil {
		api.SendBasic(c.Bot, query.Message.Chat.ID, "Something went wrong deleting your wallet details.")
		return
	}

	api.SendUpdate(c.Bot, w.NewMessageUpdate(
		"✅ Deleted",
		api.InlineKeyboard([]map[string]string{{Title: Path}}, fmt.Sprintf("user=%d", w.user.ID)),
	))
}

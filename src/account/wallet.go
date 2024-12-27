package account

import (
	"errors"
	"sync"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/util"
	"gorm.io/gorm"
)

const (
	WalletTitle = "💳 Public Wallet"
	WalletPath  = Path + "/wallet"
)

var open = sync.Map{}

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
				{"👈 Back": ".."},
			},
			PrivateOnly: true,
		},
	)
}

type Wallet struct {
	*api.Interaction[string]
	repo *repo.Repo
	user *model.User
}

func NewWallet(db *gorm.DB, query *botapi.CallbackQuery) *Wallet {
	repo := repo.NewRepo(db)
	user := &model.User{}

	if err := repo.Get(user, query.From.ID); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		user.ID = query.From.ID
	}

	return &Wallet{
		api.NewInteraction[string](query.Message, ""),
		repo,
		user,
	}
}

func OpenWallet(db *gorm.DB, query *botapi.CallbackQuery) (wallet *Wallet) {
	open.Range(func(key any, value any) bool {
		if value.(*api.Interaction[any]).Age() > time.Minute*5 {
			open.Delete(key)
		}

		return true
	})

	if data, exists := open.Load(query.From.ID); exists {
		wallet = data.(*Wallet)
	} else {
		wallet = NewWallet(db, query)
	}

	return
}

func (w *Wallet) View(c *api.Context, query *botapi.CallbackQuery) {
	util.SendBasic(c.Bot, query.Message.Chat.ID, w.user.PublicWallet)
}

func (w *Wallet) Update(c *api.Context, query *botapi.CallbackQuery) {
	w.Mutate("update", query.Message)

	msg := w.NewMessage("Okay! Send me a public wallet address to associate to your account.")
	util.SendConfig(c.Bot, &msg)

	hook := api.NewMessageHook(func(s *api.Server, m *botapi.Message, data any) {
		w = data.(*Wallet)

		if !w.Is("update") {
			return
		}

		if err := w.updateUserWallet(m.Text); err != nil {
			msg := w.NewMessage("Something went wrong updating your wallet details")
			util.SendConfig(s.Bot, &msg)
		}

		msg := w.NewMessage("✅ Saved")
		msg.ReplyMarkup = util.InlineKeyboard([]map[string]string{{
			Title: Path,
		}})

		util.SendConfig(s.Bot, &msg)
	}, w, time.Minute*5)

	c.Server.RegisterMessageHook(query.Message.Chat.ID, hook)
}

func (w *Wallet) Remove(c *api.Context, query *botapi.CallbackQuery) {
	w.user.PublicWallet = ""

	if err := w.repo.Save(w.user); err != nil {
		util.SendBasic(c.Bot, query.Message.Chat.ID, "Something went wrong deleting your wallet details.")
	} else {
		msg := w.NewMessageUpdate("✅ Deleted", util.InlineKeyboard([]map[string]string{{
			Title: Path,
		}}))
		util.SendUpdate(c.Bot, &msg)
	}
}

func (w *Wallet) updateUserWallet(addr string) (err error) {
	if !w.Is("update") {
		return
	}

	w.user.PublicWallet = addr
	err = w.repo.Save(w.user)
	return
}

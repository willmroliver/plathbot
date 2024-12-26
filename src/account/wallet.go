package account

import (
	"errors"
	"fmt"
	"sync"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/apis"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/server"
	"github.com/willmroliver/plathbot/src/util"
	"gorm.io/gorm"
)

type Wallet struct {
	*apis.Interaction[string]
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
		apis.NewInteraction[string](query.Message, ""),
		repo,
		user,
	}
}

var open = sync.Map{}

func WalletQuery(s *server.Server, query *botapi.CallbackQuery, cmd *apis.CallbackCmd) {
	open.Range(func(key any, value any) bool {
		if value.(*apis.Interaction[any]).Age() > time.Minute*5 {
			open.Delete(key)
		}

		return true
	})

	var wallet *Wallet

	if data, exists := open.Load(query.From.ID); exists {
		wallet = data.(*Wallet)
	} else {
		wallet = NewWallet(s.DB, query)
	}

	if wallet == nil {
		return
	}

	chatID := query.Message.Chat.ID

	actions := map[string]func(){
		"": func() {
			wallet.sendOptions(s.Bot, query.Message)
		},
		"view": func() {
			util.SendBasic(s.Bot, chatID, wallet.user.PublicWallet)
		},
		"update": func() {
			wallet.Mutate("update", query.Message)

			msg := wallet.NewMessage("Okay! Send me a public wallet address to associate to your account.")
			util.SendConfig(s.Bot, &msg)

			hook := server.NewMessageHook(func(s *server.Server, m *botapi.Message, data any) {
				wallet = data.(*Wallet)

				if !wallet.Is("update") {
					return
				}

				if err := wallet.UpdatePublicWallet(m.Text); err != nil {
					msg := wallet.NewMessage("Something went wrong updating your wallet details")
					util.SendConfig(s.Bot, &msg)
				}

				msg := wallet.NewMessage("✅ Saved")
				msg.ReplyMarkup = util.InlineKeyboard([]map[string]string{{
					"💻 Account": "account",
				}})

				util.SendConfig(s.Bot, &msg)
			}, wallet, time.Minute*5)

			s.RegisterMessageHook(chatID, hook)
		},
		"remove": func() {
			wallet.user.PublicWallet = ""

			if err := wallet.repo.Save(wallet.user); err != nil {
				util.SendBasic(s.Bot, chatID, "Something went wrong deleting your wallet details.")
			} else {
				msg := wallet.NewMessageUpdate("✅ Deleted", util.InlineKeyboard([]map[string]string{{
					"💻 Account": "account",
				}}))
				util.SendUpdate(s.Bot, &msg)
			}
		},
	}

	actions[cmd.Get()]()
}

func (w *Wallet) UpdatePublicWallet(addr string) (err error) {
	if !w.Is("update") {
		return
	}

	w.user.PublicWallet = addr
	err = w.repo.Save(w.user)
	return
}

func (w *Wallet) sendOptions(bot *botapi.BotAPI, message *botapi.Message) {
	options := util.InlineKeyboard([]map[string]string{
		{"✏️ Update": w.getCmd("update")},
		{"👀 View": w.getCmd("view"), "🗑️ Remove": w.getCmd("remove")},
		{"👈 Back": "account"},
	})

	msg := botapi.NewEditMessageText(message.Chat.ID, message.MessageID, "💳 Public Wallet")
	msg.ReplyMarkup = &options

	util.SendUpdate(bot, &msg)
}

func (w *Wallet) getCmd(cmd string) string {
	return fmt.Sprintf("account/wallet/%s", cmd)
}

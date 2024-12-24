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
	"github.com/willmroliver/plathbot/src/util"
	"gorm.io/gorm"
)

type Wallet struct {
	*apis.Interaction[string]
	repo *repo.Repo
	user *model.User
}

func NewWallet(query *botapi.CallbackQuery) *Wallet {
	repo := repo.NewRepo(nil)
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

func WalletQuery(bot *botapi.BotAPI, query *botapi.CallbackQuery, cmd *apis.CallbackCmd) {
	open.Range(func(key any, value any) bool {
		if value.(*apis.Interaction[any]).Age() > time.Minute*15 {
			open.Delete(key)
		}

		return true
	})

	wallet := NewWallet(query)
	if wallet == nil {
		return
	}

	actions := map[string]func(){
		"": func() {
			wallet.sendOptions(bot, query.Message)
		},
		"update": func() {
			wallet.Mutate("update", query.Message)

			msg := wallet.NewMessage("Okay! Send me a public wallet address to associate to your account.")
			util.SendConfig(bot, &msg)
		},
		"remove": func() {
			wallet.user.PublicWallet = ""

			if err := wallet.repo.Save(wallet.user); err != nil {
				util.SendBasic(bot, query.Message.Chat.ID, "Something went wrong deleting your wallet details.")
			} else {
				msg := wallet.NewMessageUpdate("✅ Done", util.InlineKeyboard([]map[string]string{{
					"Account": "account",
				}}))
				util.SendUpdate(bot, &msg)
			}
		},
	}

	actions[cmd.Get()]()
}

func (w *Wallet) sendOptions(bot *botapi.BotAPI, message *botapi.Message) {
	options := util.InlineKeyboard([]map[string]string{
		{"Update": w.getCmd("update"), "Remove": w.getCmd("remove")},
	})

	msg := botapi.NewEditMessageText(message.Chat.ID, message.MessageID, "Wallet")
	msg.ReplyMarkup = &options

	util.SendUpdate(bot, &msg)
}

func (w *Wallet) getCmd(cmd string) string {
	return fmt.Sprintf("account/wallet/%s", cmd)
}

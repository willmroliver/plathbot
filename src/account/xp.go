package account

import (
	"fmt"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/apis"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/server"
	"github.com/willmroliver/plathbot/src/util"
)

func XPQuery(s *server.Server, query *botapi.CallbackQuery, cmd *apis.CallbackCmd) {
	chatID, msgID := query.Message.Chat.ID, query.Message.MessageID
	repo := repo.NewRepo(s.DB)
	user := model.NewUser(query.From)

	if err := repo.Get(user, user.ID); err != nil {
		return
	}

	opts := util.InlineKeyboard([]map[string]string{{"👈 Back": Path}})
	msg := botapi.NewEditMessageText(chatID, msgID, fmt.Sprintf("📊 Current XP: %d", user.XP))
	msg.ReplyMarkup = &opts

	util.SendUpdate(s.Bot, &msg)
}

package account

import (
	"fmt"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/api"
	"github.com/willmroliver/plathbot/src/util"
)

const (
	XPTitle = "📈 My XP"
	XPPath  = Path + "/xp"
)

func XPQuery(c *api.Context, query *botapi.CallbackQuery, cmd *api.CallbackCmd) {
	opts := util.InlineKeyboard([]map[string]string{{"👈 Back": Path}})
	msg := botapi.NewEditMessageText(
		c.Chat.ID,
		c.Message.MessageID,
		fmt.Sprintf("📊 Current XP: %d", c.User.XP),
	)
	msg.ReplyMarkup = &opts

	util.SendUpdate(c.Bot, &msg)
}

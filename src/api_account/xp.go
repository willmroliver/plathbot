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
	msg := botapi.NewEditMessageText(
		c.Chat.ID,
		c.Message.MessageID,
		fmt.Sprintf("📊 Current XP: %d", c.GetUser().XP),
	)
	msg.ReplyMarkup = util.InlineKeyboard([]map[string]string{util.KeyboardNavRow("..")})

	util.SendUpdate(c.Bot, &msg)
}

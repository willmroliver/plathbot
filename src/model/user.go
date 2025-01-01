package model

import (
	"fmt"
	"log"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           int64                  `json:"id" gorm:"primaryKey"`
	TelegramUser *botapi.User           `json:"telegram_user" gorm:"-"`
	Username     string                 `json:"username" gorm:"size:100"`
	PublicWallet string                 `json:"public_wallet" gorm:"size:100"`
	XP           int64                  `json:"xp"`
	ReactCounts  []*ReactCount          `json:"react_counts"`
	ReactMap     map[string]*ReactCount `json:"-" gorm:"-"`
}

func NewUser(user *botapi.User) *User {
	u := &User{
		ID:           user.ID,
		TelegramUser: user,
		ReactMap:     make(map[string]*ReactCount),
	}

	u.Username = u.GetUsername()

	return u
}

func (u *User) IsAdmin(bot *botapi.BotAPI, chatID int64) bool {
	c, err := bot.GetChatMember(botapi.GetChatMemberConfig{
		ChatConfigWithUser: botapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: u.ID,
		},
	})

	if err != nil {
		log.Printf("Error getting chat config: %q", err.Error())
		return false
	}

	return c.IsAdministrator()
}

func (u *User) GetUsername() string {
	if u.Username != "" {
		return u.Username
	}

	if u.TelegramUser.UserName != "" {
		return u.TelegramUser.UserName
	}

	return fmt.Sprintf("%s_%d", u.TelegramUser.FirstName, u.ID)
}

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
	FirstName    string                 `json:"first_name" gorm:"size:64"`
	Username     string                 `json:"username" gorm:"size:100"`
	PublicWallet string                 `json:"public_wallet" gorm:"size:100"`
	ReactCounts  []*ReactCount          `json:"react_counts"`
	ReactMap     map[string]*ReactCount `json:"-" gorm:"-"`
	UserXPs      []*UserXP              `json:"user_xps"`
	UserXPMap    map[string]*UserXP     `json:"xp_map" gorm:"-"`
}

func NewUser(user *botapi.User) *User {
	u := &User{
		ID:           user.ID,
		TelegramUser: user,
		FirstName:    user.FirstName,
		ReactMap:     make(map[string]*ReactCount),
		UserXPMap:    make(map[string]*UserXP),
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

	return c.IsAdministrator() || c.IsCreator()
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

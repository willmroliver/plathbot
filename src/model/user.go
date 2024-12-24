package model

import (
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           int64        `json:"id" gorm:"primarykey"`
	TelegramUser *botapi.User `json:"telegram_user" gorm:"-"`
	PublicWallet string       `json:"public_wallet" gorm:"size:100"`
}

func NewUser(user *botapi.User) *User {
	return &User{TelegramUser: user}
}

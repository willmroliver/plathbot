package server

import (
	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type Server struct {
	Bot *botapi.BotAPI
	DB  *gorm.DB
}

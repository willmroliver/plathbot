package service

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"gorm.io/gorm"
)

const (
	XPTitleEngage = "💕 Engage XP"
	XPTitleGames  = "🎮 Games XP"
)

var xpServices = map[*gorm.DB]*UserXPService{}

type UserXPService struct {
	UserRepo   *repo.UserRepo
	UserXPRepo *repo.UserXPRepo
}

func NewUserXPService(db *gorm.DB) *UserXPService {
	if s, ok := xpServices[db]; ok {
		return s
	}

	s := &UserXPService{
		UserRepo:   repo.NewUserRepo(db),
		UserXPRepo: repo.NewUserXPRepo(db),
	}

	xpServices[db] = s
	return s
}

func (s *UserXPService) UpdateXPs(user *tgbotapi.User, title string, points int64) (err error) {
	u := s.UserRepo.Get(user)
	if u == nil || points == 0 || title == "" {
		err = fmt.Errorf("invalid args: (user = %v, title = %q, points = %d)", user, title, points)
		log.Printf("UpdateXPs() - %q", err.Error())
		return
	}

	xp := u.UserXPMap[title]
	if xp == nil {
		xp = model.NewUserXP(title, u.ID)
		u.UserXPMap[title] = xp
	}

	err = s.UserXPRepo.ShiftXP(xp, points)
	return
}

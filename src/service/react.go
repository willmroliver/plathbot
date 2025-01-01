package service

import (
	"errors"
	"fmt"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"gorm.io/gorm"
)

type ReactService struct {
	userRepo  *repo.UserRepo
	reactRepo *repo.ReactRepo
	countRepo *repo.ReactCountRepo
}

func NewReactService(db *gorm.DB) *ReactService {
	return &ReactService{
		userRepo:  repo.NewUserRepo(db),
		reactRepo: repo.NewReactRepo(db),
		countRepo: repo.NewReactCountRepo(db),
	}
}

func (r *ReactService) Untrack(emoji string) (err error) {
	if err = r.reactRepo.Delete(emoji); err != nil {
		return
	}

	repo.OnUserCache(func(u *model.User) bool {
		delete(u.ReactMap, emoji)
		return true
	})

	r.countRepo.DeleteBy(&model.ReactCount{}, "emoji", emoji)
	return
}

func (r *ReactService) UpdateCounts(m *botapi.Message) (err error) {
	if m.User == nil {
		err = errors.New("user is nil")
		return
	}

	if len(m.OldReaction)+len(m.NewReaction) == 0 {
		return
	}

	user := r.userRepo.Get(m.User)

	if user == nil {
		err = fmt.Errorf("cannot find user %+v", user)
		return
	}

	for _, react := range m.OldReaction {
		if react == nil || react.Emoji == "" || r.reactRepo.Get(react.Emoji) == nil {
			continue
		}

		if data := user.ReactMap[react.Emoji]; data != nil && data.Count > 0 {
			if err = r.countRepo.ShiftCount(data, -1); err != nil {
				return
			}
		}
	}

	for _, react := range m.NewReaction {
		if react == nil || react.Emoji == "" || r.reactRepo.Get(react.Emoji) == nil {
			continue
		}

		if data := user.ReactMap[react.Emoji]; data != nil {
			if err = r.countRepo.ShiftCount(data, 1); err != nil {
				return
			}
		} else {
			data = model.NewReactCount(react.Emoji, user.ID)
			if err = r.countRepo.ShiftCount(data, 1); err != nil {
				return
			}

			user.ReactMap[react.Emoji] = data
		}
	}

	return
}

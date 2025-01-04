package service

import (
	"errors"
	"fmt"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/util"
	"gorm.io/gorm"
)

var reactServices = map[*gorm.DB]*ReactService{}

type ReactService struct {
	UserRepo  *repo.UserRepo
	ReactRepo *repo.ReactRepo
	CountRepo *repo.ReactCountRepo
}

func NewReactService(db *gorm.DB) *ReactService {
	if s, ok := reactServices[db]; ok {
		return s
	}

	s := &ReactService{
		UserRepo:  repo.NewUserRepo(db),
		ReactRepo: repo.NewReactRepo(db),
		CountRepo: repo.NewReactCountRepo(db),
	}

	reactServices[db] = s
	return s
}

func (r *ReactService) Untrack(emoji string) (err error) {
	if err = r.ReactRepo.Delete(emoji); err != nil {
		return
	}

	repo.OnUserCache(func(u *model.User) bool {
		delete(u.ReactMap, emoji)
		return true
	})

	r.CountRepo.DeleteBy(&model.ReactCount{}, "emoji", emoji)
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

	user := r.UserRepo.Get(m.User)

	if user == nil {
		err = fmt.Errorf("cannot find user %+v", user)
		return
	}

	for _, react := range m.OldReaction {
		if react == nil {
			continue
		}

		react.Emoji = util.NormalizeEmoji(react.Emoji)
		if react.Emoji == "" || r.ReactRepo.Get(react.Emoji) == nil {
			continue
		}

		if data := user.ReactMap[react.Emoji]; data != nil && data.Count > 0 {
			if err = r.CountRepo.ShiftCount(data, -1); err != nil {
				return
			}
		}
	}

	for _, react := range m.NewReaction {
		if react == nil {
			continue
		}

		react.Emoji = util.NormalizeEmoji(react.Emoji)
		if react.Emoji == "" || r.ReactRepo.Get(react.Emoji) == nil {
			continue
		}

		if data := user.ReactMap[react.Emoji]; data != nil {
			if err = r.CountRepo.ShiftCount(data, 1); err != nil {
				return
			}
		} else {
			data = model.NewReactCount(react.Emoji, user.ID)
			if err = r.CountRepo.ShiftCount(data, 1); err != nil {
				return
			}

			user.ReactMap[react.Emoji] = data
		}
	}

	return
}

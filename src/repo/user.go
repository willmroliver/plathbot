package repo

import (
	"errors"
	"log"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/ds"
	"github.com/willmroliver/plathbot/src/model"
	"gorm.io/gorm"
)

var cache = ds.NewLRUCache[int64, *model.User](100)

type UserRepo struct {
	*Repo
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		Repo: NewRepo(db),
	}
}

func (r *UserRepo) Get(u *botapi.User) *model.User {
	if user, ok := cache.Load(u.ID); ok && user != nil {
		return user
	}

	user := model.NewUser(u)
	if user == nil {
		return nil
	}

	if err := r.db.Preload("ReactCounts").First(user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error reading user %d record: %q", user.ID, err.Error())
		return nil
	}

	for _, count := range user.ReactCounts {
		user.ReactMap[count.ID] = count
	}

	return user
}

func (r *UserRepo) Save(user *model.User) (err error) {
	if user == nil {
		return
	}

	user.ReactCounts = make([]*model.ReactCount, len(user.ReactMap))
	i := 0

	for _, count := range user.ReactMap {
		user.ReactCounts[i] = count
		i++
	}

	err = r.Repo.Save(user)
	return
}

func (r *UserRepo) ShiftXP(u *botapi.User, xp int64) (err error) {
	if user := r.Get(u); user != nil {
		user.XP += xp

		if err = r.Save(user); err != nil {
			log.Printf("Error updating user %d record: %q", user.ID, err.Error())
		}
	}

	return
}

func (r *UserRepo) UpdateWallet(u *botapi.User, addr string) (err error) {
	if user := r.Get(u); user != nil {
		user.PublicWallet = addr

		if err = r.Save(user); err != nil {
			log.Printf("Error updating user %d record: %q", user.ID, err.Error())
		}
	}

	return
}

func (r *UserRepo) UpdateReacts(m *botapi.Message) (err error) {
	if m.User == nil || len(m.OldReaction)+len(m.NewReaction) == 0 {
		return
	}

	user := r.Get(m.User)

	if user == nil {
		return
	}

	for _, react := range m.OldReaction {
		if react == nil || react.Emoji == "" {
			continue
		}

		if data := user.ReactMap[react.Emoji]; data != nil && data.Count > 0 {
			data.Count -= 1
			if err = r.Repo.Save(data); err != nil {
				return
			}
		}
	}

	for _, react := range m.NewReaction {
		if react == nil || react.Emoji == "" {
			continue
		}

		if data := user.ReactMap[react.Emoji]; data != nil {
			data.Count += 1
			if err = r.Repo.Save(data); err != nil {
				return
			}
		} else {
			data = &model.ReactCount{
				ID:     react.Emoji,
				UserID: user.ID,
				Count:  1,
			}

			user.ReactMap[react.Emoji] = data
			if err = r.Repo.Save(data); err != nil {
				return
			}
		}
	}

	return
}

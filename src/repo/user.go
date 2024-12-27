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

	if err := r.db.First(user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error reading user %d record: %q", user.ID, err.Error())
		return nil
	}

	return user
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

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

func OnUserCache(cb func(*model.User) bool) {
	cache.ForEach(cb)
}

type UserRepo struct {
	*Repo
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		NewRepo(db),
	}
}

func (r *UserRepo) Get(u *botapi.User) *model.User {
	cache.Lock()
	defer cache.Unlock()

	if user, ok := cache.Load(u.ID); ok && user != nil {
		return user
	}

	user := model.NewUser(u)
	if user == nil {
		return nil
	}

	if err := r.db.Preload("ReactCounts").Preload("UserXPs").First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.Save(user)
		} else {
			log.Printf("Error reading user %d record: %q", user.ID, err.Error())
			return nil
		}
	}

	initUser(user)
	cache.Save(user.ID, user)
	return user
}

func (r *UserRepo) AllWhere(clause string, conditions ...interface{}) (users []*model.User) {
	users = []*model.User{}

	err := r.Repo.db.
		Preload("ReactCounts").
		Preload("UserXPs").
		Where(clause, conditions...).
		Find(&users).
		Error

	if err != nil {
		log.Printf("Error reading users: %q", err.Error())
		return nil
	}

	for _, u := range users {
		initUser(u)
	}

	return
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

	user.UserXPs = make([]*model.UserXP, len(user.UserXPMap))
	i = 0

	for _, xp := range user.UserXPMap {
		user.UserXPs[i] = xp
		i++
	}

	err = r.Repo.Save(user)
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

func (r *UserRepo) AllRedditUsernames() (users []string) {
	r.db.Model(&model.User{}).Where("reddit_username IS NOT NULL").Pluck("reddit_username", &users)
	return
}

func initUser(u *model.User) {
	u.ReactMap = make(map[string]*model.ReactCount)
	u.UserXPMap = make(map[string]*model.UserXP)

	for _, count := range u.ReactCounts {
		u.ReactMap[count.Emoji] = count
	}

	for _, xp := range u.UserXPs {
		u.UserXPMap[xp.Title] = xp
	}
}

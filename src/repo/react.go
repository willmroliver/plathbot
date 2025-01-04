package repo

import (
	"sync"

	"github.com/willmroliver/plathbot/src/model"
	"gorm.io/gorm"
)

var (
	reacts    map[string]*model.React = nil
	reactsMux                         = &sync.RWMutex{}
)

type ReactRepo struct {
	*Repo
}

func NewReactRepo(db *gorm.DB) *ReactRepo {
	return &ReactRepo{
		Repo: NewRepo(db),
	}
}

func (r *ReactRepo) Save(emoji, title string) (err error) {
	if reacts == nil {
		r.All()
	}

	react := &model.React{Emoji: emoji, Title: title}
	if err = r.Repo.Save(react); err == nil {
		reactsMux.Lock()
		defer reactsMux.Unlock()

		reacts[emoji] = react
	}

	return
}

func (r *ReactRepo) Delete(emoji string) (err error) {
	react := &model.React{}
	if err = r.Repo.DeleteBy(react, "emoji", emoji); err == nil {
		reactsMux.Lock()
		defer reactsMux.Unlock()

		delete(reacts, emoji)
	}

	return
}

func (r *ReactRepo) All() (results []*model.React) {
	if reacts != nil {
		reactsMux.RLock()
		defer reactsMux.RUnlock()

		results = make([]*model.React, len(reacts))
		i := 0

		for _, react := range reacts {
			results[i] = react
			i++
		}

		return
	}

	results = make([]*model.React, 0)
	if err := r.Repo.All(&results); err != nil {
		return
	}

	reactsMux.Lock()
	defer reactsMux.Unlock()

	reacts = make(map[string]*model.React)

	for _, react := range results {
		reacts[react.Emoji] = react
	}

	return
}

func (r *ReactRepo) Get(emoji string) *model.React {
	if reacts == nil {
		r.All()
	}

	reactsMux.RLock()
	defer reactsMux.RUnlock()

	if react, ok := reacts[emoji]; ok {
		return react
	}

	return nil
}

package repo

import (
	"github.com/willmroliver/plathbot/src/model"
	"gorm.io/gorm"
)

var (
	reacts = map[string]*model.React{}
	loaded = false
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
	react := &model.React{Emoji: emoji, Title: title}
	if err = r.Repo.Save(react); err == nil {
		reacts[emoji] = react
	}

	return
}

func (r *ReactRepo) Delete(emoji string) (err error) {
	react := &model.React{}
	if err = r.Repo.DeleteBy(react, "emoji", emoji); err == nil {
		delete(reacts, emoji)
	}

	return
}

func (r *ReactRepo) All() (results []*model.React) {
	if loaded {
		results = make([]*model.React, len(reacts))
		i := 0

		for _, react := range reacts {
			results[i] = react
			i++
		}

		return results
	}

	results = make([]*model.React, 0)
	if err := r.Repo.All(&results); err != nil {
		return
	}

	for _, react := range results {
		reacts[react.Emoji] = react
	}

	loaded = true
	return
}

func (r *ReactRepo) Get(emoji string) *model.React {
	if !loaded {
		r.All()
	}

	if react, ok := reacts[emoji]; ok {
		return react
	}

	return nil
}

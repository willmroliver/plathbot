package repo

import (
	"time"

	"github.com/willmroliver/plathbot/src/model"
	"gorm.io/gorm"
)

type RedditPostRepo struct {
	*Repo
}

func NewRedditPostRepo(db *gorm.DB) *RedditPostRepo {
	return &RedditPostRepo{
		Repo: NewRepo(db),
	}
}

func (r *RedditPostRepo) All() []*model.RedditPost {
	posts := []*model.RedditPost{}

	// Delete associated comments and posts
	r.Repo.db.Where("expires_at < ?", time.Now()).Delete(&model.RedditPost{})

	if err := r.Repo.All(&posts); err != nil {
		return nil
	}

	return posts
}

func (r *RedditPostRepo) Expired() (posts []*model.RedditPost) {
	r.db.Preload("Comments").Where("expires_at < ?", time.Now()).Find(&posts)
	return
}

func (r *RedditPostRepo) Delete(postIDs ...string) (err error) {
	err = r.db.Select("Comments").Where("post_id IN ?", postIDs).Delete(&model.RedditPost{}).Error
	return
}

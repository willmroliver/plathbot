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

func (r *RedditPostRepo) Delete(postID string) (err error) {
	err = r.db.Select("Comments").Delete(&model.RedditPost{PostID: postID}).Error
	return
}

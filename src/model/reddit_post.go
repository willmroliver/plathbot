package model

import (
	"time"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type RedditPost struct {
	PostID    string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp"`
	ExpiresAt time.Time `json:"expires_at" gorm:"type:timestamp"`
}

func NewRedditPost(post *reddit.Post) *RedditPost {
	return &RedditPost{
		PostID:    post.ID,
		Title:     post.Title,
		URL:       post.URL,
		CreatedAt: post.Created.Time,
	}
}

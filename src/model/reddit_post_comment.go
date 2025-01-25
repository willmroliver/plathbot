//go:build reddit
// +build reddit

package model

import "github.com/vartanbeno/go-reddit/v2/reddit"

type RedditPostComment struct {
	PostID   string `json:"post_id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"primaryKey"`
	User     *User  `json:"user" gorm:"foreignKey:RedditUsername;references:Username"`
	Comment  string `json:"comment" gorm:"default:null;type:varchar(50)"`
}

func NewRedditPostComment(postID string, c *reddit.Comment) *RedditPostComment {
	comment := c.Body
	if len(comment) > 50 {
		comment = comment[:50]
	}

	return &RedditPostComment{
		PostID:   postID,
		Username: c.Author,
		Comment:  comment,
	}
}

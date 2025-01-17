package service_test

import (
	"os"
	"testing"
	"time"

	"github.com/willmroliver/plathbot/src/db"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/service"
)

func TestAll(t *testing.T) {
	users := []*model.User{
		{Username: "test", RedditUsername: "test"},
		{Username: "test2", RedditUsername: "test2"},
		{Username: "test3", RedditUsername: "test3"},
	}

	post := &model.RedditPost{
		PostID:    "123",
		Title:     "Test Post",
		URL:       "https://www.reddit.com/r/test",
		ExpiresAt: time.Now().Add(-time.Second),
	}

	comments := []*model.RedditPostComment{
		{
			PostID:   "123",
			Username: users[0].RedditUsername,
			Comment:  "Test comment",
		},
		{
			PostID:   "123",
			Username: users[1].RedditUsername,
			Comment:  "Test comment 2",
		},
		{
			PostID:   "123",
			Username: users[2].RedditUsername,
			Comment:  "Test comment 3",
		},
	}

	conn, _ := db.Open(os.Getenv("TEST_DB_NAME"))
	r := repo.NewRedditPostRepo(conn)

	if err := r.Save(post); err != nil {
		t.Errorf("RedditPostRepo Save() - Unexpected error: %q", err.Error())
		return
	}

	defer r.Delete(post.PostID)

	for _, u := range users {
		if err := r.Repo.Save(u); err != nil {
			t.Errorf("UserRepo Save() - Unexpected error: %q", err.Error())
			return
		} else {
			defer r.Repo.Delete(u)
		}
	}

	for _, c := range comments {
		if err := r.Save(c); err != nil {
			t.Errorf("RedditPostRepo SaveComment() - Unexpected error: %q", err.Error())
			return
		} else {
			defer r.Repo.Delete(c)
		}
	}

	s := service.NewRedditService(conn)
	s.All()

	xps := s.UserXPService.UserXPRepo.List(service.XPTitleReddit, "", 0, 100)

	if len(xps) != 3 {
		t.Errorf("RedditService All() - Expected 3 XP records, got %d", len(xps))
		return
	}

	for i, xp := range xps {
		if xp.UserID != users[i].ID {
			t.Errorf("RedditService All() - Expected username %q, got %q", users[i].ID, xp.UserID)
			return
		}
		if xp.XP != 1 {
			t.Errorf("RedditService All() - Expected 1 XP, got %d", xp.XP)
			return
		}
	}
}

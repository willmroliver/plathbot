package service

import (
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"gorm.io/gorm"
)

type RedditService struct {
	RedditPostRepo *repo.RedditPostRepo
	UserXPService  *UserXPService
}

func NewRedditService(db *gorm.DB) *RedditService {
	return &RedditService{
		repo.NewRedditPostRepo(db),
		NewUserXPService(db),
	}
}

func (s *RedditService) All() (posts []*model.RedditPost) {
	expired := s.RedditPostRepo.Expired()

	ids := make([]string, len(expired))
	for i, p := range expired {
		ids[i] = p.PostID
	}

	s.RedditPostRepo.Delete(ids...)

	users := make(map[string]struct{})

	for _, p := range expired {
		for _, c := range p.Comments {
			users[c.Username] = struct{}{}
		}
	}

	usernames, i := make([]string, len(users)), 0
	for u := range users {
		usernames[i] = u
		i++
	}

	s.UserXPService.UpdateXPsWhere(XPTitleReddit, 1, "reddit_username IN ?", usernames)

	return s.RedditPostRepo.All()
}

package reddit

import (
	"log"
	"sync"
	"time"

	goreddit "github.com/vartanbeno/go-reddit/v2/reddit"
	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/service"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TrackPosts(db *gorm.DB, freq time.Duration) func() {
	run := true
	cancel := func() {
		run = false
	}

	userRepo := repo.NewUserRepo(db)
	redditService := service.NewRedditService(db)

	go func() {
		for run {
			time.Sleep(freq)

			posts := redditService.All()
			if len(posts) == 0 {
				continue
			}

			users := userRepo.AllRedditUsernames()
			if len(users) == 0 {
				continue
			}

			puMap := make(map[string]map[string]struct{}, len(users))

			for _, post := range posts {
				puMap[post.PostID] = make(map[string]struct{}, len(users))
				for _, u := range users {
					puMap[post.PostID][u] = struct{}{}
				}
			}

			ch, wg := make(chan *model.RedditPostComment, len(posts)*len(users)), sync.WaitGroup{}
			wg.Add(len(posts))

			for postID, users := range puMap {
				go PollComments(postID, 0, -time.Second, func(p *goreddit.PostAndComments, _ any) (done bool) {
					done = true

					for _, c := range p.Comments {
						if _, ok := users[c.Author]; ok {
							ch <- model.NewRedditPostComment(p.Post.ID, c)
						}
					}

					wg.Done()
					return
				}, nil)

				time.Sleep(time.Millisecond * 100)
			}

			wg.Wait()

			data := make([]*model.RedditPostComment, len(ch))
			for i := 0; i < len(data); i++ {
				data[i] = <-ch
			}

			if len(data) == 0 {
				continue
			}

			err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "post_id"}, {Name: "username"}},
				DoUpdates: clause.AssignmentColumns([]string{"comment"}),
			}).Create(&data).Error

			if err != nil {
				log.Printf("Error saving comments: %q", err.Error())
			}
		}
	}()

	return cancel
}

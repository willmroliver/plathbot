package reddit

import (
	"context"
	"log"
	"time"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

var client *reddit.Client

func Client() *reddit.Client {
	if client == nil {
		var err error

		client, err = reddit.NewClient(reddit.Credentials{}, reddit.FromEnv, reddit.WithUserAgent("p1ath_bot:XbJRexGWkV2SAPUsqrZWGA (by u/p1ath_bot)"))

		if err != nil {
			log.Printf("NewClient() - %q", err.Error())
			return nil
		}

		return client
	}

	return client
}

func PostBasic(title, text string) *reddit.Submitted {
	post, _, err := Client().Post.SubmitText(newContext(), reddit.SubmitTextRequest{
		Subreddit: "p1ath_bot",
		Title:     title,
		Text:      text,
	})

	if err != nil {
		log.Printf("PostBasic() - %q", err.Error())
		return nil
	}

	return post
}

func PollComments(postID string, every, dur time.Duration, cb func(*reddit.PostAndComments, any) bool, payload any) []*reddit.Comment {
	data, _, err := Client().Post.Get(newContext(), postID)

	if err != nil {
		log.Printf("PollComments() - %q", err.Error())
		return nil
	}

	until := time.Now().Add(dur)

	for {
		if cb(data, payload) || time.Now().After(until) {
			break
		}

		time.Sleep(every)

		data, _, err = Client().Post.Get(newContext(), postID)
		if err != nil {
			log.Printf("PollComments() - %q", err.Error())
			return nil
		}
	}

	return data.Comments
}

func newContext() context.Context {
	c, _ := context.WithTimeout(context.Background(), time.Second*10)
	return c
}

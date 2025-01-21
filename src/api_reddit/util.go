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

		client, err = reddit.NewClient(reddit.Credentials{}, reddit.FromEnv)

		if err != nil {
			log.Printf("NewClient() - %q", err.Error())
			return nil
		}

		return client
	}

	return client
}

func PostBasic(title, text string) *reddit.Submitted {
	ctx, cancel := newContext()
	defer cancel()

	post, _, err := Client().Post.SubmitText(ctx, reddit.SubmitTextRequest{
		Subreddit: "p1ath_bot_messages",
		Title:     title,
		Text:      text,
	})

	if err != nil {
		log.Printf("PostBasic() - %q", err.Error())
		return nil
	}

	return post
}

func GetPost(id string) *reddit.Post {
	ctx, cancel := newContext()
	defer cancel()

	post, _, err := Client().Post.Get(ctx, id)

	if err != nil {
		log.Printf("GetPost() - %q", err.Error())
		return nil
	}

	return post.Post
}

func DeletePost(fullID string) {
	ctx, cancel := newContext()
	defer cancel()

	_, err := Client().Post.Delete(ctx, fullID)

	if err != nil {
		log.Printf("DeletePost() - %q", err.Error())
	}
}

func LockPost(fullID string) {
	ctx, cancel := newContext()
	defer cancel()

	_, err := Client().Post.Lock(ctx, fullID)

	if err != nil {
		log.Printf("LockPost() - %q", err.Error())
	}
}

func HidePost(fullID string) {
	ctx, cancel := newContext()
	defer cancel()

	_, err := Client().Post.Hide(ctx, fullID)

	if err != nil {
		log.Printf("HidePost() - %q", err.Error())
	}
}

func PollComments(postID string, every, dur time.Duration, cb func(*reddit.PostAndComments, any) bool, payload any) []*reddit.Comment {
	ctx, cancel := newContext()
	defer cancel()

	data, _, err := Client().Post.Get(ctx, postID)

	if err != nil {
		log.Printf("PollComments() - %q", err.Error())
		return nil
	}

	until := time.Now().Add(dur)

	for !cb(data, payload) && time.Now().Before(until) {
		time.Sleep(every)

		ctx1, cancel1 := newContext()
		data, _, err = Client().Post.Get(ctx1, postID)
		cancel1()

		if err != nil {
			log.Printf("PollComments() - %q", err.Error())
			return nil
		}
	}

	return data.Comments
}

func PollInbox(every, dur time.Duration, cb func([]*reddit.Message, []*reddit.Message, any) bool, payload any) []*reddit.Message {
	if dur == 0 {
		dur = every - time.Millisecond
	}

	until, after := time.Now().Add(dur), ""

	for time.Now().Before(until) {
		ctx, cancel := newContext()
		defer cancel()

		coms, msgs, _, err := Client().Message.InboxUnread(ctx, &reddit.ListOptions{After: after})

		if err != nil {
			log.Printf("PollInbox() - %q", err.Error())
			return nil
		}

		if cb(coms, msgs, payload) {
			return append(coms, msgs...)
		}

		time.Sleep(every)
	}

	return nil
}

func newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*10)
}

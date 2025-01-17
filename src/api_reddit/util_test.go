package reddit_test

import (
	"testing"

	"github.com/joho/godotenv"
	reddit "github.com/willmroliver/plathbot/src/api_reddit"
)

func TestGetPost(t *testing.T) {
	godotenv.Load("./../../.env")

	if post := reddit.GetPost("1i27cue"); post == nil {
		t.Error("GetPost() - expected post, got nil")
	}
}

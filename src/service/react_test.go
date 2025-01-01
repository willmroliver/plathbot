package service_test

import (
	"os"
	"testing"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/db"
	"github.com/willmroliver/plathbot/src/repo"
	"github.com/willmroliver/plathbot/src/service"
)

const (
	FireEmoji  = `🔥`
	SmileEmoji = `😁`
)

func TestReactCount(t *testing.T) {
	tgUser := &botapi.User{
		ID: 1,
	}

	addReact := &botapi.Message{
		User:        tgUser,
		OldReaction: []*botapi.ReactionType{},
		NewReaction: []*botapi.ReactionType{
			{Type: "emoji", Emoji: FireEmoji},
		},
	}
	changeReact := &botapi.Message{
		User: tgUser,
		OldReaction: []*botapi.ReactionType{
			{Type: "emoji", Emoji: FireEmoji},
		},
		NewReaction: []*botapi.ReactionType{
			{Type: "emoji", Emoji: SmileEmoji},
		},
	}
	removeReact := &botapi.Message{
		User: tgUser,
		OldReaction: []*botapi.ReactionType{
			{Type: "emoji", Emoji: SmileEmoji},
		},
		NewReaction: []*botapi.ReactionType{},
	}

	conn, _ := db.Open(os.Getenv("TEST_DB_NAME"))
	service := service.NewReactService(conn)
	userRepo := repo.NewUserRepo(conn)
	user := userRepo.Get(tgUser)

	// Add React
	if err := service.UpdateCounts(addReact); err != nil {
		t.Errorf("UpdatedReacts() - Unexpected error: %q", err.Error())
	}

	if count, ok := user.ReactMap[FireEmoji]; !ok || count == nil {
		t.Errorf("ReactMap[%s] - Expected exists; Got %v, %v", FireEmoji, ok, count)
	}

	// Change React
	if err := service.UpdateCounts(changeReact); err != nil {
		t.Errorf("UpdatedReacts() - Unexpected error: %q", err.Error())
	}

	if count, ok := user.ReactMap[FireEmoji]; ok && count.Count > 0 {
		t.Errorf("ReactMap[%s] - Expected falsey; Got %v, %v", FireEmoji, ok, count)
	}

	if count, ok := user.ReactMap[SmileEmoji]; !ok || count == nil {
		t.Errorf("ReactMap[%s] - Expected exists; Got %v, %v", SmileEmoji, ok, count)
	}

	// Remove React
	if err := service.UpdateCounts(removeReact); err != nil {
		t.Errorf("UpdatedReacts() - Unexpected error: %q", err.Error())
	}

	if count, ok := user.ReactMap[SmileEmoji]; ok && count.Count > 0 {
		t.Errorf("ReactMap[%s] - Expected falsey; Got %v, %v", SmileEmoji, ok, count)
	}
}

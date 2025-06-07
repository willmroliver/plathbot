package service_test

import (
	"os"
	"testing"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/db"
	"github.com/willmroliver/plathbot/src/service"
	"github.com/willmroliver/plathbot/src/util"
)

func TestUserXP(t *testing.T) {
	tgUser := &botapi.User{
		ID: 1,
	}

	conn, _ := db.Open(os.Getenv("TEST_DB_NAME"))
	db.Migrate(conn)

	s := service.NewUserXPService(conn)
	user := s.UserRepo.Get(tgUser)
	conn.Exec("DELETE FROM user_xps WHERE title = ?", service.XPTitleEngage)

	if err := s.UpdateXPs(tgUser, service.XPTitleEngage, 50); err != nil {
		t.Errorf("UpdateXPs() - Unexpected error: %q", err.Error())
	}

	if count, ok := user.UserXPMap[service.XPTitleEngage]; !ok || count == nil || count.XP != 50 {
		t.Errorf("UserXPMap[%q] - Expected exists, %d; Got %v, %v", service.XPTitleEngage, 50, ok, count)
	}

	past := time.Now().AddDate(0, -1, 0)
	user.UserXPMap[service.XPTitleEngage].WeekFrom = util.LastMonday(&past)

	if err := s.UpdateXPs(tgUser, service.XPTitleEngage, 50); err != nil {
		t.Errorf("UpdateXPs() - Unexpected error: %q", err.Error())
	}

	if count, ok := user.UserXPMap[service.XPTitleEngage]; !ok || count == nil || count.XP != 100 {
		t.Errorf("UserXPMap[%q] - Expected exists, %d; Got %v, %v", service.XPTitleEngage, 100, ok, count)
	} else if count.WeekXP != 50 {
		t.Errorf("WeekXP - Expected %d; Got %d", 50, count.WeekXP)
	}
}

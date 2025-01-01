package repo

import (
	"time"

	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/util"
	"gorm.io/gorm"
)

type ReactCountRepo struct {
	*Repo
}

func NewReactCountRepo(db *gorm.DB) *ReactCountRepo {
	return &ReactCountRepo{
		Repo: NewRepo(db),
	}
}

func (r *ReactCountRepo) ShiftCount(react *model.ReactCount, count int) (err error) {
	if react == nil || count == 0 {
		return
	}

	now := time.Now()

	if monday := util.LastMonday(&now); react.WeekFrom.Compare(monday) != 0 {
		react.WeekCount = 0
		react.WeekFrom = monday
	}

	if first := util.FirstOfMonth(&now); react.MonthFrom.Compare(first) != 0 {
		react.MonthCount = 0
		react.MonthFrom = first
	}

	shift := func(n, by int) int {
		if c := n + by; c >= 0 {
			return c
		}

		return 0
	}

	react.Count = shift(react.Count, count)
	react.WeekCount = shift(react.WeekCount, count)
	react.MonthCount = shift(react.MonthCount, count)

	err = r.Save(react)
	return
}

func (r *ReactCountRepo) TopCount(emoji string) (c *model.ReactCount) {
	c = &model.ReactCount{Emoji: emoji}

	if err := r.TopBy(c, "count DESC"); err != nil {
		return nil
	}

	return
}

func (r *ReactCountRepo) TopWeekCount(emoji string) (c *model.ReactCount) {
	c = &model.ReactCount{Emoji: emoji}

	if err := r.TopBy(c, "count_week DESC"); err != nil {
		return nil
	}

	return
}

func (r *ReactCountRepo) TopMonthCount(emoji string) (c *model.ReactCount) {
	c = &model.ReactCount{Emoji: emoji}

	if err := r.TopBy(c, "count_month DESC"); err != nil {
		return nil
	}

	return
}

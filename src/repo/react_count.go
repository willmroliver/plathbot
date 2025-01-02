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

func (r *ReactCountRepo) List(emoji, order string, offset, lim int) (counts []*model.ReactCount) {
	counts = make([]*model.ReactCount, 0)

	if order != "" {
		r.db.Order(order)
	}

	if emoji != "" {
		r.db.Where("emoji = ?", emoji)
	}

	r.db.Offset(offset).Limit(lim).Find(&counts)
	return
}

// TopCounts returns the all-time highest count & user for each tracked emoji
func (r *ReactCountRepo) TopCounts() (c []*model.ReactCount) {
	c = make([]*model.ReactCount, 0)

	if err := r.db.Raw(`
		WITH top_counts AS (
			SELECT 
				emoji,
				user_id,
				count,
				ROW_NUMBER() OVER (PARTITION BY emoji ORDER BY count DESC, user_id ASC) as rn
			FROM react_counts
		)
		SELECT emoji, user_id, count
		FROM top_counts
		WHERE rn = 1
	`).Preload("User").Find(&c).Error; err != nil {
		return nil
	}

	return
}

// TopMonthly returns the all-time highest count & user for each tracked emoji
//
// The `Count` field is populated with the MonthCount value to support code-homogeneity
func (r *ReactCountRepo) TopMonthly() (c []*model.ReactCount) {
	c = make([]*model.ReactCount, 0)

	if err := r.db.Raw(`
		WITH top_counts AS (
			SELECT 
				emoji,
				user_id,
				month_count AS count,
				ROW_NUMBER() OVER (PARTITION BY emoji ORDER BY month_count DESC, user_id ASC) as rn
			FROM react_counts
		)
		SELECT emoji, user_id, count
		FROM top_counts
		WHERE rn = 1
	`).Preload("User").Find(&c).Error; err != nil {
		return nil
	}

	for _, count := range c {
		if count != nil {
			count.MonthCount = count.Count
		}
	}

	return
}

// TopWeekly returns the all-time highest count & user for each tracked emoji
//
// The `Count` field is populated with the WeekCount value to support code-homogeneity
func (r *ReactCountRepo) TopWeekly() (c []*model.ReactCount) {
	c = make([]*model.ReactCount, 0)

	if err := r.db.Raw(`
		WITH top_counts AS (
			SELECT 
				emoji,
				user_id,
				week_count AS count,
				ROW_NUMBER() OVER (PARTITION BY emoji ORDER BY week_count DESC, user_id ASC) as rn
			FROM react_counts
		)
		SELECT emoji, user_id, count
		FROM top_counts
		WHERE rn = 1
	`).Preload("User").Find(&c).Error; err != nil {
		return nil
	}

	for _, count := range c {
		if count != nil {
			count.WeekCount = count.Count
		}
	}

	return
}

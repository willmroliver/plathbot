package repo

import (
	"time"

	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/util"
	"gorm.io/gorm"
)

type UserXPRepo struct {
	*Repo
}

func NewUserXPRepo(db *gorm.DB) *UserXPRepo {
	return &UserXPRepo{
		Repo: NewRepo(db),
	}
}

func (r *UserXPRepo) ShiftXP(xp *model.UserXP, points int64) (err error) {
	if xp == nil || points == 0 {
		return
	}

	now := time.Now()

	if monday := util.LastMonday(&now); xp.WeekFrom.Compare(monday) != 0 {
		xp.WeekXP = 0
		xp.WeekFrom = monday
	}

	if first := util.FirstOfMonth(&now); xp.MonthFrom.Compare(first) != 0 {
		xp.MonthXP = 0
		xp.MonthFrom = first
	}

	shift := func(n, by int64) int64 {
		if c := n + by; c >= 0 {
			return c
		}

		return 0
	}

	xp.XP = shift(xp.XP, points)
	xp.WeekXP = shift(xp.WeekXP, points)
	xp.MonthXP = shift(xp.MonthXP, points)

	err = r.Save(xp)
	return
}

func (r *UserXPRepo) List(title, order string, offset, lim int) (xps []*model.UserXP) {
	xps = make([]*model.UserXP, 0)

	if order != "" {
		r.db.Order(order)
	}

	if title != "" {
		r.db.Where("title = ?", title)
	}

	r.db.Offset(offset).Limit(lim).Find(&xps)
	return
}

// TopXPs returns the all-time highest count & user for each tracked title
func (r *UserXPRepo) TopXPs(title, order string, offset, limit int) (c []*model.UserXP) {
	c = make([]*model.UserXP, 0)

	err := r.db.
		Where("title = ?", title).
		Joins("User").
		Offset(offset).
		Limit(limit).
		Find(&c).
		Error

	if err != nil {
		return nil
	}

	return
}

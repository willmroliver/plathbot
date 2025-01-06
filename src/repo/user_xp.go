package repo

import (
	"sync"
	"time"

	"github.com/willmroliver/plathbot/src/model"
	"github.com/willmroliver/plathbot/src/util"
	"gorm.io/gorm"
)

var (
	xpTitles    map[string]bool = nil
	xpTitlesMux                 = &sync.Mutex{}
)

type UserXPRepo struct {
	*Repo
}

func NewUserXPRepo(db *gorm.DB) *UserXPRepo {
	return &UserXPRepo{
		NewRepo(db),
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

	if err = r.Save(xp); err != nil {
		xpTitlesMux.Lock()
		defer xpTitlesMux.Unlock()

		xpTitles[xp.Title] = true
	}

	return
}

func (r *UserXPRepo) Titles() (titles []string) {
	xpTitlesMux.Lock()
	defer xpTitlesMux.Unlock()

	if xpTitles != nil {
		titles = make([]string, len(xpTitles))
		i := 0

		for title := range xpTitles {
			titles[i] = title
			i++
		}

		return titles
	}

	xpTitles = make(map[string]bool)
	titles = make([]string, 0)

	if err := r.db.Model(&model.UserXP{}).Distinct("title").Pluck("title", &titles).Error; err != nil {
		return nil
	}

	for _, title := range titles {
		xpTitles[title] = true
	}

	return
}

func (r *UserXPRepo) List(title, order string, offset, lim int) (xps []*model.UserXP) {
	xps = make([]*model.UserXP, 0)

	query := r.db.Offset(offset).Limit(lim)

	if order != "" {
		query.Order(order)
	}

	if title != "" {
		query.Where("title = ?", title)
	}

	query.Find(&xps)
	return
}

// TopXPs returns the all-time highest count & user for each tracked title
func (r *UserXPRepo) TopXPs(title, order string, offset, limit int) (c []*model.UserXP) {
	c = make([]*model.UserXP, 0)

	err := r.db.
		Where("title = ?", title).
		Joins("User").
		Order(order).
		Offset(offset).
		Limit(limit).
		Find(&c).
		Error

	if err != nil {
		return nil
	}

	return
}

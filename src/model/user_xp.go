package model

import (
	"time"

	"github.com/willmroliver/plathbot/src/util"
)

type UserXP struct {
	Title     string    `json:"title" gorm:"primaryKey;size:50"`
	UserID    int64     `json:"user_id" gorm:"primaryKey"`
	User      *User     `json:"user" gorm:"foreignKey:UserID;references:ID"`
	XP        int64     `json:"xp"`
	WeekXP    int64     `json:"week_xp"`
	MonthXP   int64     `json:"month_xp"`
	WeekFrom  time.Time `json:"week_from" gorm:"type:date"`
	MonthFrom time.Time `json:"month_from" gorm:"type:date"`
}

func NewUserXP(title string, userID int64) *UserXP {
	now := time.Now()

	return &UserXP{
		Title:     title,
		UserID:    userID,
		WeekFrom:  util.LastMonday(&now),
		MonthFrom: util.FirstOfMonth(&now),
	}
}

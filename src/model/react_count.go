package model

import (
	"time"

	"github.com/willmroliver/plathbot/src/util"
)

type ReactCount struct {
	Emoji      string    `json:"emoji" gorm:"primaryKey;type:char(4)"`
	UserID     int64     `json:"user_id" gorm:"primaryKey"`
	User       *User     `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Count      int       `json:"count"`
	WeekCount  int       `json:"week_count"`
	MonthCount int       `json:"month_count"`
	WeekFrom   time.Time `json:"week_from" gorm:"type:date"`
	MonthFrom  time.Time `json:"month_from" gorm:"type:date"`
}

func NewReactCount(emoji string, userID int64) *ReactCount {
	now := time.Now()

	return &ReactCount{
		Emoji:     emoji,
		UserID:    userID,
		WeekFrom:  util.LastMonday(&now),
		MonthFrom: util.FirstOfMonth(&now),
	}
}

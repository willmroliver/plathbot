package model

import (
	"gorm.io/gorm"
)

type ReactCount struct {
	*gorm.Model
	ID     string `json:"id" gorm:"primarykey;type:char(4)"`
	UserID int64  `json:"user_id"`
	Count  int    `json:"count"`
}

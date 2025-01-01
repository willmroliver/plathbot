package model

type React struct {
	Emoji string `json:"emoji" gorm:"primaryKey;type:char(4)"`
	Title string `json:"title" gorm:"size:50"`
}

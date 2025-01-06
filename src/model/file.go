package model

import (
	"log"
	"strings"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type File struct {
	FileUniqueID string `json:"file_unique_id" gorm:"primaryKey;type:text"`
	FileID       string `json:"file_id" gorm:"type:text"`
	Name         string `json:"name" gorm:"type:text;unique;index"`
	Tags         string `json:"tags" gorm:"type:text"`
}

func NewFile(file any, name string, tags ...string) (f *File) {
	var id, uid string

	switch v := file.(type) {
	case *botapi.File:
		id, uid = v.FileID, v.FileUniqueID
		tags = append(tags, "File")
	case *botapi.PhotoSize:
		id, uid = v.FileID, v.FileUniqueID
		tags = append(tags, "Photo")
	default:
		log.Printf("Type: %T", v)
		return
	}

	f = &File{
		uid,
		id,
		name,
		strings.Join(tags, ","),
	}
	return
}

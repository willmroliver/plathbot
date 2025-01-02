package db

import (
	"log"
	"reflect"

	"github.com/willmroliver/plathbot/src/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) (err error) {
	tables := []any{
		&model.User{},
		&model.UserXP{},
		&model.ReactCount{},
		&model.React{},
	}

	for _, table := range tables {
		if err = db.AutoMigrate(table); err != nil {
			log.Printf("Error migrating %s: %q", reflect.TypeOf(table).Elem().Name(), err.Error())
			return
		}
	}

	return
}

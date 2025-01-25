package db

import (
	"log"
	"reflect"

	"github.com/willmroliver/plathbot/src/model"
	"gorm.io/gorm"
)

var tables = []any{
	&model.File{},
	&model.User{},
	&model.UserXP{},
	&model.React{},
	&model.ReactCount{},
}

func MigrateModel(table any) {
	tables = append(tables, table)
}

func Migrate(db *gorm.DB) (err error) {
	for _, table := range tables {
		if err = db.AutoMigrate(table); err != nil {
			log.Printf("Error migrating %s: %q", reflect.TypeOf(table).Elem().Name(), err.Error())
			return
		}
	}

	return
}

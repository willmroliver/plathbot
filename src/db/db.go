package db

import (
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var pool = map[string]*gorm.DB{}
var mutex = sync.Mutex{}

func Open(name string) (*gorm.DB, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if db, ok := pool[name]; ok {
		return db, nil
	}

	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	pool[name] = db
	return db, nil
}

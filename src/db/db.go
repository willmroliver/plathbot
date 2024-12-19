package db

import (
	"log"
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

	log.Println("SQLite connection opened. Migrating tables...")

	if err = Migrate(db); err != nil {
		log.Printf("Migration error")
		return nil, err
	} else {
		log.Println("Migration complete.")
	}

	pool[name] = db
	return db, nil
}

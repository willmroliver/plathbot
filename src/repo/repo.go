package repo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	d "github.com/willmroliver/plathbot/src/db"
	"gorm.io/gorm"
)

var (
	repos    = map[*gorm.DB]*Repo{}
	reposMux = &sync.Mutex{}
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	if db == nil {
		if conn, err := d.Open(os.Getenv("DB_NAME")); err != nil {
			panic(fmt.Sprintf("Error opening 'test.db': %q", err.Error()))
		} else {
			db = conn
		}
	}

	reposMux.Lock()
	defer reposMux.Unlock()

	if r := repos[db]; r != nil {
		return r
	}

	r := &Repo{db}
	repos[db] = r
	return r
}

func (r *Repo) Save(m any) (err error) {
	if err = r.db.Save(m).Error; err != nil {
		log.Printf("Repo Save() error: %q", err.Error())
	}

	return
}

func (r *Repo) DeleteBy(m any, col string, val any) (err error) {
	if err = r.db.Where(col+" = ?", val).Delete(m).Error; err != nil {
		log.Printf("Repo Delete() error: %q", err.Error())
	}

	return
}

func (r *Repo) Delete(m any) (err error) {
	if err = r.db.Delete(m).Error; err != nil {
		log.Printf("Repo Delete() error: %q", err.Error())
	}

	return
}

func (r *Repo) GetBy(m any, col string, val any) (err error) {
	err = r.db.Where(col+" = ?", val).First(m).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Repo Get() error: %q", err.Error())
		return
	}

	return nil
}

func (r *Repo) Get(m any) (err error) {
	err = r.db.First(m).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Repo Get() error: %q", err.Error())
		return
	}

	return nil
}

func (r *Repo) TopBy(m any, order string) (err error) {
	err = r.db.Order(order).Take(m).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Repo Top() error: %q", err.Error())
		return nil
	}

	return
}

func (r *Repo) All(m any) (err error) {
	if err = r.db.Find(m).Error; err != nil {
		log.Printf("Repo All() error: %q", err.Error())
	}

	return
}

func (r *Repo) AllWhere(m any, clause string, args ...any) (err error) {
	if err = r.db.Where(clause, args...).Find(m).Error; err != nil {
		log.Printf("Repo AllWhere() error: %q", err.Error())
	}

	return
}

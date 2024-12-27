package repo

import (
	"errors"
	"fmt"
	"log"

	d "github.com/willmroliver/plathbot/src/db"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	if db == nil {
		conn, err := d.Open("test.db")
		if err != nil {
			panic(fmt.Sprintf("Error opening 'test.db': %q", err.Error()))
		}

		return &Repo{db: conn}
	}

	return &Repo{db}
}

func (r *Repo) Save(m any) (err error) {
	if err = r.db.Save(m).Error; err != nil {
		log.Printf("Repo Save() error: %q", err.Error())
	}

	return
}

func (r *Repo) GetBy(m any, col string, val any) (err error) {
	err = r.db.Where(fmt.Sprintf("%s = ?", col), val).First(m).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Repo Get() error: %q", err.Error())
		return
	}

	return nil
}

func (r *Repo) Get(m any, id any) error {
	return r.GetBy(m, "id", id)
}

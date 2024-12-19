package repo

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db}
}

func (r *Repo) Save(m any) (err error) {
	if err = r.db.Save(m).Error; err != nil {
		log.Printf("Repo Save() error: %q", err.Error())
	}

	return
}

func (r *Repo) GetWhere(m any, col string, val any) (err error) {
	if err = r.db.Where(fmt.Sprintf("%s = ?", col), val).First(m).Error; err != nil {
		log.Printf("Repo Get() error: %q", err.Error())
	}

	return
}

func (r *Repo) Get(m any, id any) error {
	return r.GetWhere(m, "id", id)
}

package repo

import (
	"errors"
	"fmt"
	"log"
	"os"

	d "github.com/willmroliver/plathbot/src/db"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	if db == nil {
		conn, err := d.Open(os.Getenv("DB_NAME"))
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

func (r *Repo) DeleteBy(m any, col string, val any) (err error) {
	if err = r.db.Where(fmt.Sprintf("%s = ?", col), val).Delete(m).Error; err != nil {
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
	err = r.db.Where(fmt.Sprintf("%s = ?", col), val).First(m).Error

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

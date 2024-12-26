package repo

import (
	"errors"
	"log"

	"github.com/willmroliver/plathbot/src/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	*Repo
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		Repo: NewRepo(db),
	}
}

func (r *UserRepo) ShiftXP(userID, xp int64) (err error) {
	err = r.db.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("xp", gorm.Expr("xp + ?", xp)).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error updating user %d record: %q", userID, err.Error())
	}

	return
}

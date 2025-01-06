package repo

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/willmroliver/plathbot/src/model"
)

type FileRepo struct {
	*Repo
}

func NewFileRepo(db *gorm.DB) *FileRepo {
	return &FileRepo{
		NewRepo(db),
	}
}

func (r *FileRepo) Save(f any, name string, tags ...string) (file *model.File) {
	if file = model.NewFile(f, name, tags...); file == nil {
		return
	}

	err := r.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "file_unique_id"}},             // Unique column(s)
			DoUpdates: clause.AssignmentColumns([]string{"file_id", "name"}), // Columns to update
		},
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},                                 // Unique column(s)
			DoUpdates: clause.AssignmentColumns([]string{"file_unique_id", "file_id"}), // Columns to update
		},
	).Create(file).Error

	if err != nil {
		log.Printf("NewFileRepo Save() - %q", err.Error())
		return nil
	}

	return
}

func (r *FileRepo) Get(name string) (file *model.File) {
	file = &model.File{}
	if r.Repo.GetBy("name", name, file) != nil {
		return nil
	}

	return
}

func (r *FileRepo) List(path, order string, offset, lim int, tags ...string) (files []*model.File) {
	files = make([]*model.File, 0)
	query := r.db.Where("name LIKE ?", path+"%")

	for _, t := range tags {
		query.Where("tags LIKE ?", "%"+t+"%")
	}

	if order != "" {
		query.Order(order)
	}

	query.Offset(offset).Limit(lim).Find(&files)

	return
}

func (r *FileRepo) Delete(name string) error {
	return r.Repo.DeleteBy(&model.File{}, "name", name)
}

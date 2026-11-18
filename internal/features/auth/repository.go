package auth

import (
	"gorm.io/gorm"

	"app/internal/database"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(u *database.User) error { return r.db.Create(u).Error }

func (r *Repository) ByUsername(name string) (*database.User, error) {
	var u database.User
	return &u, r.db.Where("username = ?", name).First(&u).Error
}

func (r *Repository) ByID(id uint) (*database.User, error) {
	var u database.User
	return &u, r.db.First(&u, id).Error
}

func (r *Repository) Exists(username, email string, exceptID uint) (bool, error) {
	var n int64
	err := r.db.Model(&database.User{}).
		Where("(username = ? OR email = ?) AND id <> ?", username, email, exceptID).Count(&n).Error
	return n > 0, err
}

func (r *Repository) Save(u *database.User) error { return r.db.Save(u).Error }

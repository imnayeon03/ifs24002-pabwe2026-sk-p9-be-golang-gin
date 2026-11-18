package post

import (
	"strings"

	"gorm.io/gorm"

	"app/internal/database"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Public(f Filter) ([]database.Post, error) {
	q := r.db.Preload("User").Where("is_published = ?", true).Order("published_at desc, id desc")
	if f.Query != "" {
		like := "%" + strings.ToLower(f.Query) + "%"
		q = q.Where("LOWER(title) LIKE ? OR LOWER(excerpt) LIKE ? OR LOWER(content) LIKE ?", like, like, like)
	}
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	posts := []database.Post{}
	err := q.Find(&posts).Error
	return posts, err
}

func (r *Repository) PublicBySlug(slug string) (*database.Post, error) {
	var p database.Post
	return &p, r.db.Preload("User").Where("is_published = ? AND slug = ?", true, slug).First(&p).Error
}

func (r *Repository) ByUser(userID uint) ([]database.Post, error) {
	posts := []database.Post{}
	err := r.db.Preload("User").Where("user_id = ?", userID).Order("id desc").Find(&posts).Error
	return posts, err
}

func (r *Repository) All() ([]database.Post, error) {
	posts := []database.Post{}
	err := r.db.Preload("User").Order("id desc").Find(&posts).Error
	return posts, err
}

func (r *Repository) Mine(userID, id uint) (*database.Post, error) {
	var p database.Post
	return &p, r.db.Preload("User").Where("user_id = ?", userID).First(&p, id).Error
}

func (r *Repository) SlugTaken(slug string, exceptID uint) (bool, error) {
	var n int64
	err := r.db.Model(&database.Post{}).Where("slug = ? AND id <> ?", slug, exceptID).Count(&n).Error
	return n > 0, err
}

func (r *Repository) Save(p *database.Post) error { return r.db.Save(p).Error }

func (r *Repository) Delete(p *database.Post) error { return r.db.Delete(p).Error }

func (r *Repository) Count(userID uint, onlyPublished bool) (int64, error) {
	var n int64
	q := r.db.Model(&database.Post{}).Where("user_id = ?", userID)
	if onlyPublished {
		q = q.Where("is_published = ?", true)
	}
	err := q.Count(&n).Error
	return n, err
}

func (r *Repository) Users() ([]database.User, error) {
	var users []database.User
	err := r.db.Order("id").Find(&users).Error
	return users, err
}

package database

import (
	"time"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"app/internal/config"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:255;uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:text;not null" json:"-"`
	Role      string    `gorm:"size:50;not null;default:member" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Post struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	User           *User      `gorm:"constraint:OnDelete:CASCADE" json:"author,omitempty"`
	Title          string     `gorm:"size:255;not null" json:"title"`
	Slug           string     `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Category       string     `gorm:"size:100;not null;default:Umum" json:"category"`
	Excerpt        string     `gorm:"type:text" json:"excerpt"`
	Content        string     `gorm:"type:text" json:"content"`
	CoverImagePath string     `gorm:"type:text" json:"cover_image_path"`
	IsPublished    bool       `gorm:"not null;default:false" json:"is_published"`
	PublishedAt    *time.Time `json:"published_at"`
	ReadingTime    int        `gorm:"-" json:"reading_time"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Connect membuka PostgreSQL untuk runtime.
func Connect(cfg config.Config) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
}

// ConnectMemory membuka SQLite in-memory untuk pengujian mandiri.
func ConnectMemory() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1) // satu koneksi agar memori dibagi
	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Post{})
}

func seedUser(db *gorm.DB, username, password, role string) error {
	var n int64
	if err := db.Model(&User{}).Where("username = ?", username).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return err
	}
	return db.Create(&User{Username: username, Email: username + "@delcom.org", Password: string(hash), Role: role}).Error
}

// Seed membuat akun bawaan admin/admin dan member/member.
func Seed(db *gorm.DB, cfg config.Config) error {
	if err := seedUser(db, cfg.AdminUsername, cfg.AdminPassword, "admin"); err != nil {
		return err
	}
	return seedUser(db, "member", "member", "member")
}

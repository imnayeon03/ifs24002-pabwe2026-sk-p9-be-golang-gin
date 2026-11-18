package post

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gorm.io/gorm"

	"app/internal/database"
)

var (
	ErrNotFound        = errors.New("artikel tidak ditemukan")
	ErrInvalidCategory = errors.New("kategori tidak valid")
)

type UserSummary struct {
	database.User
	TotalPosts     int64 `json:"total_posts"`
	TotalPublished int64 `json:"total_published"`
}

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// Slugify: huruf kecil, non-alfanumerik menjadi tanda hubung.
func Slugify(title string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(title) {
		if r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "artikel"
	}
	return slug
}

// ReadingTime = ceil(kata / 200), minimal 1 menit.
func ReadingTime(content string) int {
	words := len(strings.Fields(content))
	if t := int(math.Ceil(float64(words) / 200.0)); t > 1 {
		return t
	}
	return 1
}

func wrap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func withTime(p *database.Post) *database.Post {
	p.ReadingTime = ReadingTime(p.Content)
	return p
}

func withTimes(posts []database.Post) []database.Post {
	for i := range posts {
		withTime(&posts[i])
	}
	return posts
}

func (s *Service) uniqueSlug(title string, exceptID uint) (string, error) {
	base := Slugify(title)
	slug := base
	for i := 2; ; i++ {
		taken, err := s.repo.SlugTaken(slug, exceptID)
		if err != nil {
			return "", err
		}
		if !taken {
			return slug, nil
		}
		slug = base + "-" + strconv.Itoa(i)
	}
}

func validCategory(c string) (string, error) {
	if c == "" {
		return "Umum", nil
	}
	for _, k := range Categories {
		if k == c {
			return c, nil
		}
	}
	return "", ErrInvalidCategory
}

func (s *Service) Public(f Filter) ([]database.Post, error) {
	posts, err := s.repo.Public(f)
	return withTimes(posts), err
}

func (s *Service) PublicBySlug(slug string) (*database.Post, error) {
	p, err := s.repo.PublicBySlug(slug)
	if err != nil {
		return nil, wrap(err)
	}
	return withTime(p), nil
}

func (s *Service) Mine(userID uint) ([]database.Post, error) {
	posts, err := s.repo.ByUser(userID)
	return withTimes(posts), err
}

func (s *Service) Get(userID, id uint) (*database.Post, error) {
	p, err := s.repo.Mine(userID, id)
	if err != nil {
		return nil, wrap(err)
	}
	return withTime(p), nil
}

func (s *Service) Create(userID uint, req PostRequest) (*database.Post, error) {
	cat, err := validCategory(req.Category)
	if err != nil {
		return nil, err
	}
	slug, err := s.uniqueSlug(req.Title, 0)
	if err != nil {
		return nil, err
	}
	p := &database.Post{UserID: userID, Title: req.Title, Slug: slug, Category: cat,
		Excerpt: req.Excerpt, Content: req.Content, CoverImagePath: req.CoverImagePath}
	if err := s.repo.Save(p); err != nil {
		return nil, err
	}
	return withTime(p), nil
}

func (s *Service) Update(userID, id uint, req PostRequest) (*database.Post, error) {
	p, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	cat, err := validCategory(req.Category)
	if err != nil {
		return nil, err
	}
	p.Title, p.Category, p.Excerpt, p.Content, p.CoverImagePath = req.Title, cat, req.Excerpt, req.Content, req.CoverImagePath
	p.User = nil
	if err := s.repo.Save(p); err != nil {
		return nil, err
	}
	return withTime(p), nil
}

func (s *Service) TogglePublish(userID, id uint) (*database.Post, error) {
	p, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	p.IsPublished = !p.IsPublished
	if p.IsPublished {
		now := time.Now()
		p.PublishedAt = &now
	} else {
		p.PublishedAt = nil
	}
	p.User = nil
	if err := s.repo.Save(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(userID, id uint) error {
	p, err := s.Get(userID, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(p)
}

func (s *Service) Stats(userID uint) (Stats, error) {
	total, err := s.repo.Count(userID, false)
	if err != nil {
		return Stats{}, err
	}
	pub, err := s.repo.Count(userID, true)
	if err != nil {
		return Stats{}, err
	}
	return Stats{Total: total, TotalTerbit: pub, BelumSelesai: total - pub}, nil
}

func (s *Service) AdminUsers() ([]UserSummary, error) {
	users, err := s.repo.Users()
	if err != nil {
		return nil, err
	}
	out := make([]UserSummary, 0, len(users))
	for _, u := range users {
		total, err := s.repo.Count(u.ID, false)
		if err != nil {
			return nil, err
		}
		pub, err := s.repo.Count(u.ID, true)
		if err != nil {
			return nil, err
		}
		out = append(out, UserSummary{User: u, TotalPosts: total, TotalPublished: pub})
	}
	return out, nil
}

func (s *Service) AdminPosts() ([]database.Post, error) {
	posts, err := s.repo.All()
	return withTimes(posts), err
}

func (s *Service) AdminUserPosts(userID uint) ([]database.Post, error) { return s.Mine(userID) }

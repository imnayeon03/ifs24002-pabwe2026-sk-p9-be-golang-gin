package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"app/internal/database"
)

var (
	ErrDuplicate   = errors.New("username atau email sudah terdaftar")
	ErrCredentials = errors.New("username atau password salah")
	ErrNotFound    = errors.New("pengguna tidak ditemukan")
)

type Service struct {
	repo   *Repository
	secret string
}

func NewService(repo *Repository, secret string) *Service {
	return &Service{repo: repo, secret: secret}
}

func (s *Service) Register(req RegisterRequest) (*database.User, error) {
	dup, err := s.repo.Exists(req.Username, req.Email, 0)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, ErrDuplicate
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &database.User{Username: req.Username, Email: req.Email, Password: string(hash), Role: "member"}
	return u, s.repo.Create(u)
}

func (s *Service) Login(req LoginRequest) (string, *database.User, error) {
	u, err := s.repo.ByUsername(req.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return "", nil, ErrCredentials
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": u.ID, "username": u.Username, "role": u.Role,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}).SignedString([]byte(s.secret))
	return token, u, err
}

func (s *Service) Profile(id uint) (*database.User, error) {
	u, err := s.repo.ByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *Service) UpdateProfile(id uint, req UpdateProfileRequest) (*database.User, error) {
	u, err := s.Profile(id)
	if err != nil {
		return nil, err
	}
	if req.Username != "" {
		u.Username = req.Username
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	dup, err := s.repo.Exists(u.Username, u.Email, id)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, ErrDuplicate
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		u.Password = string(hash)
	}
	return u, s.repo.Save(u)
}

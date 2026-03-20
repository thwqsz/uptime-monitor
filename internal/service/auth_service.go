package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewAuthService(repo repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, email, password string) error {
	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return ErrUserAlreadyExists
	} else if errors.Is(err, sql.ErrNoRows) {
		hash, err := HashPassword(password)
		if err != nil {
			return err
		}
		user := models.User{Email: email, PasswordHash: hash}
		err = s.repo.CreateUser(ctx, &user)
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	if err := ComparePassword(user.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}
	token, err := GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return "", err
	}
	return token, nil
}

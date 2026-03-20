package service

import (
	"context"
	"errors"
	"strings"

	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
)

var ErrInvalidURL = errors.New("invalid url")
var ErrInvalidTimeout = errors.New("invalid timeout")
var ErrInvalidInterval = errors.New("invalid interval")

type TargetService struct {
	repo repository.TargetRepository
}

func NewTargetService(repo repository.TargetRepository) *TargetService {
	return &TargetService{
		repo: repo,
	}
}

func (s *TargetService) CreateTarget(ctx context.Context, userID int64, url string, timeout, interval int) (*models.Target, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http") {
		return nil, ErrInvalidURL
	}
	if interval <= 0 {
		return nil, ErrInvalidInterval
	}
	if timeout <= 0 {
		return nil, ErrInvalidTimeout
	}

	target := models.Target{
		UserID:       userID,
		URL:          url,
		Timeout:      timeout,
		IntervalTime: interval,
	}
	err := s.repo.CreateTarget(ctx, &target)
	if err != nil {
		return nil, err
	}
	return &target, nil
}

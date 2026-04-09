package service

import (
	"context"
	"errors"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
	neturl "net/url"
	"strings"
)

var ErrInvalidURL = errors.New("invalid url")
var ErrInvalidTimeout = errors.New("invalid timeout")
var ErrInvalidInterval = errors.New("invalid interval")
var ErrInvalidUserID = errors.New("invalid userID")
var ErrNoTargetFound = errors.New("no target found")
var ErrMultipleTargetsDeleted = errors.New("multiple target deleted")
var ErrInvalidTargetID = errors.New("invalid targetID")

type SchedulerControl interface {
	StartTarget(target *models.Target)
	StopTarget(targetID int64)
}
type TargetService struct {
	repo      repository.TargetRepository
	scheduler SchedulerControl
}

func NewTargetService(repo repository.TargetRepository, scheduler SchedulerControl) *TargetService {
	return &TargetService{
		repo:      repo,
		scheduler: scheduler,
	}
}

func (s *TargetService) CreateTarget(ctx context.Context, userID int64, url string, timeout, interval int) (*models.Target, error) {
	url = strings.TrimSpace(url)
	if _, err := neturl.ParseRequestURI(url); err != nil {
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
	if s.scheduler != nil {
		s.scheduler.StartTarget(&target)
	}
	return &target, nil
}

func (s *TargetService) ListTargets(ctx context.Context, userID int64) ([]*models.Target, error) {
	if userID < 1 {
		return nil, ErrInvalidUserID
	}
	targets, err := s.repo.GetTargetsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return targets, nil
}

func (s *TargetService) DeleteTarget(ctx context.Context, targetID, userID int64) error {
	if userID < 1 {
		return ErrInvalidUserID
	}
	if targetID < 1 {
		return ErrInvalidTargetID
	}
	rowsAffected, err := s.repo.DeleteTarget(ctx, targetID, userID)
	if err != nil {
		return err
	}
	switch rowsAffected {
	case 0:
		return ErrNoTargetFound
	case 1:
		if s.scheduler != nil {
			s.scheduler.StopTarget(targetID)
		}
		return nil
	default:
		return ErrMultipleTargetsDeleted
	}
}

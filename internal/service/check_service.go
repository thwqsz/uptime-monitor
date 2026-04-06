package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/thwqsz/uptime-monitor/internal/checker"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
)

var ErrAccessDenied = errors.New("access denied")

type CheckService struct {
	checkRepo  repository.CheckLogRepository
	targetRepo repository.TargetRepository
	check      checker.Checker
}

func NewCheckService(checkRepo repository.CheckLogRepository, targetRepo repository.TargetRepository, check checker.Checker) *CheckService {
	return &CheckService{
		check:      check,
		targetRepo: targetRepo,
		checkRepo:  checkRepo,
	}
}

func (s *CheckService) runCheckTarget(ctx context.Context, target *models.Target) (*models.CheckLog, error) {
	timeout := time.Duration(target.Timeout) * time.Second
	resCheck, err := s.check.Check(ctx, target.URL, timeout)
	if err != nil {
		return nil, err
	}

	respTimeInt := int(resCheck.Duration.Milliseconds())

	var status string
	if resCheck.StatusCode/100 == 2 && resCheck.Error == nil {
		status = "up"
	} else {
		status = "down"
	}
	var errorMsg *string
	var ans models.CheckLog
	if resCheck.Error != nil {
		msg := fmt.Sprint(resCheck.Error)
		errorMsg = &msg
	}
	ans = models.CheckLog{
		TargetID:       target.ID,
		StatusCode:     resCheck.StatusCode,
		ResponseTimeMs: respTimeInt,
		ErrorMsg:       errorMsg,
		Status:         status,
	}
	err = s.checkRepo.CreateCheckLog(ctx, &ans)
	if err != nil {
		return nil, err
	}
	return &ans, nil
}

func (s *CheckService) CheckTargetSystem(ctx context.Context, targetID int64) (*models.CheckLog, error) {
	target, err := s.targetRepo.GetTargetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoTargetFound
		}
		return nil, err
	}
	return s.runCheckTarget(ctx, target)
}

func (s *CheckService) CheckTargetForUser(ctx context.Context, targetID int64, userID int64) (*models.CheckLog, error) {
	target, err := s.targetRepo.GetTargetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoTargetFound
		}
		return nil, err
	}
	if target.UserID != userID {
		return nil, ErrAccessDenied
	}
	check, err := s.runCheckTarget(ctx, target)
	if err != nil {
		return nil, err
	}
	return check, nil
}

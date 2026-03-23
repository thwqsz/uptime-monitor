package service

import (
	"context"
	"fmt"
	"time"

	"github.com/thwqsz/uptime-monitor/internal/checker"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
)

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

func (s *CheckService) CheckTarget(ctx context.Context, targetID int64) (*models.CheckLog, error) {
	if targetID < 1 {
		return nil, ErrInvalidTargetID
	}
	target, err := s.targetRepo.GetTargetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
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
		TargetID:       targetID,
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

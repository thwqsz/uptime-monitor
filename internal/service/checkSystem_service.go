package service

import (
	"context"
	"errors"
	"time"

	"github.com/thwqsz/uptime-monitor/internal/contracts"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
)

type CheckServiceForKafka struct {
	checkRepo repository.CheckLogRepository
}

func NewCheckServiceForKafka(checkRepo repository.CheckLogRepository) *CheckServiceForKafka {
	return &CheckServiceForKafka{
		checkRepo: checkRepo,
	}
}

func (s *CheckServiceForKafka) ProcessCheckResult(ctx context.Context, resCheck *contracts.CheckResult) error {
	if resCheck.TargetID < 1 {
		return errors.New("invalid targetID")
	}
	if !(resCheck.Status == "up" || resCheck.Status == "down") {
		return errors.New("unexpected status")
	}
	var errMsg *string
	if resCheck.ErrorMsg == "" {
		errMsg = nil
	} else {
		errMsg = &resCheck.ErrorMsg
	}
	checkedAt, err := time.Parse(time.RFC3339, resCheck.CheckedAt)
	if err != nil {
		return err
	}
	ans := models.CheckLog{
		TargetID:       resCheck.TargetID,
		StatusCode:     resCheck.StatusCode,
		ResponseTimeMs: resCheck.ResponseTimeMs,
		ErrorMsg:       errMsg,
		Status:         resCheck.Status,
		CheckedAt:      checkedAt,
	}
	err = s.checkRepo.CreateCheckLog(ctx, &ans)
	if err != nil {
		return err
	}
	return nil
}

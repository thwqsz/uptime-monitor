package service

import (
	"context"
	"errors"
	"time"

	"github.com/thwqsz/uptime-monitor/internal/contracts"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
)

type TaskSender interface {
	SendTask(ctx context.Context, target *models.Target) (string, error)
}

type CheckServiceForKafka struct {
	checkRepo repository.CheckLogRepository
}

func NewCheckServiceForKafka(checkRepo repository.CheckLogRepository) *CheckServiceForKafka {
	return &CheckServiceForKafka{
		checkRepo: checkRepo,
	}
}

type CheckBeforeSend struct {
	targetRepo  repository.TargetRepository
	sendToKafka TaskSender
}

func NewCheckBeforeSend(targetRepo repository.TargetRepository, sendToKafka TaskSender) *CheckBeforeSend {
	return &CheckBeforeSend{
		targetRepo:  targetRepo,
		sendToKafka: sendToKafka,
	}
}

func (s *CheckServiceForKafka) ProcessCheckResultForSystem(ctx context.Context, resCheck *contracts.CheckResult) error {
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

//func (s *CheckBeforeSend) SendCheckForUser(ctx context.Context, targetID int64, userID int64) (string, error) {
//	target, err := s.targetRepo.GetTargetByID(ctx, targetID)
//	if err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return "", ErrNoTargetFound
//		}
//		return "", err
//	}
//	if target.UserID != userID {
//		return "", ErrAccessDenied
//	}
//	uniqID, err := s.sendToKafka.SendTask(ctx, target)
//	if err != nil {
//		return "", err
//	}
//	return uniqID, nil
//}

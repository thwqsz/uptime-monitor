package service

import (
	"context"
	"errors"
	"time"

	"github.com/thwqsz/uptime-monitor/internal/contracts"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
	"go.uber.org/zap"
)

type TaskSender interface {
	SendTask(ctx context.Context, target *models.Target) (string, error)
}

type CacheManager interface {
	SaveStatus(ctx context.Context, targetID int64, status string) error
	GetLastStatus(ctx context.Context, id int64) (string, error)
}

type CheckServiceForKafka struct {
	checkRepo    repository.CheckLogRepository
	cacheManager CacheManager
	log          *zap.Logger
}

func NewCheckServiceForKafka(checkRepo repository.CheckLogRepository, cacheManager CacheManager, log *zap.Logger) *CheckServiceForKafka {
	return &CheckServiceForKafka{
		checkRepo:    checkRepo,
		cacheManager: cacheManager,
		log:          log,
	}
}

//type CheckBeforeSend struct {
//	targetRepo  repository.TargetRepository
//	sendToKafka TaskSender
//}
//
//func NewCheckBeforeSend(targetRepo repository.TargetRepository, sendToKafka TaskSender) *CheckBeforeSend {
//	return &CheckBeforeSend{
//		targetRepo:  targetRepo,
//		sendToKafka: sendToKafka,
//	}
//}

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
	prevStatus, err := s.cacheManager.GetLastStatus(ctx, resCheck.TargetID)
	if err != nil {
		return err
	}
	if prevStatus == "up" && resCheck.Status == "down" {
		s.log.Info("server is down")
	}
	if prevStatus == "down" && resCheck.Status == "up" {
		s.log.Info("server is recovered")
	}
	
	return s.cacheManager.SaveStatus(ctx, resCheck.TargetID, resCheck.Status)
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

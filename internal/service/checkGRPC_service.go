package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/thwqsz/uptime-monitor/internal/grpc/checkerpb"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/repository"
	"google.golang.org/grpc"
)

type GRPCChecker interface {
	Check(ctx context.Context, in *checkerpb.CheckRequest, opts ...grpc.CallOption) (*checkerpb.CheckResponse, error)
}

type ManualCheckService struct {
	targetRepo repository.TargetRepository
	logRepo    repository.CheckLogRepository
	checkGRPC  GRPCChecker
}

func NewManualService(targetRepo repository.TargetRepository, logRepo repository.CheckLogRepository, checkGRPC GRPCChecker) *ManualCheckService {
	return &ManualCheckService{targetRepo: targetRepo, logRepo: logRepo, checkGRPC: checkGRPC}
}

func (s *ManualCheckService) ManualCheck(ctx context.Context, targetID, userID int64) (*models.CheckLog, error) {
	target, err := s.getOwnedTarget(ctx, targetID, userID)
	if err != nil {
		return nil, err
	}
	in := checkerpb.CheckRequest{
		Url:        target.URL,
		TimeoutSec: int32(target.Timeout),
	}
	resp, err := s.checkGRPC.Check(ctx, &in)
	if err != nil {
		return nil, err
	}
	var errMsg *string
	if resp.ErrMsg == "" {
		errMsg = nil
	} else {
		errMsg = &resp.ErrMsg
	}
	now := time.Now().UTC()
	log := &models.CheckLog{
		TargetID:       targetID,
		StatusCode:     int(resp.StatusCode),
		Status:         resp.Status,
		ErrorMsg:       errMsg,
		ResponseTimeMs: int(resp.ResponseTimeMs),
		CheckedAt:      now,
	}
	err = s.logRepo.CreateCheckLog(ctx, log)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func (s *ManualCheckService) getOwnedTarget(ctx context.Context, targetID, userID int64) (*models.Target, error) {
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
	return target, nil
}

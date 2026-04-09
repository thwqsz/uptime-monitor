package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/thwqsz/uptime-monitor/internal/checker"
	"github.com/thwqsz/uptime-monitor/internal/mocks"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/service"
)

func TestCheckService_CheckTargetSystem_404CreateDownLog(t *testing.T) {
	ctx := context.Background()
	targetID := int64(1)

	targetRepo := mocks.NewTargetRepository(t)
	checkLogRepo := mocks.NewCheckLogRepository(t)
	checkerMock := mocks.NewChecker(t)

	svc := service.NewCheckService(checkLogRepo, targetRepo, checkerMock)

	target := &models.Target{
		ID:      targetID,
		Timeout: 5,
		URL:     "https://examlpe.com",
	}

	targetRepo.On("GetTargetByID", mock.Anything, targetID).Return(target, nil)
	checkLogRepo.On("CreateCheckLog", mock.Anything, mock.MatchedBy(func(log *models.CheckLog) bool {
		return log.StatusCode == 404 && log.TargetID == targetID && log.Status == "down" && log.ErrorMsg == nil
	})).Return(nil)
	checkerMock.On("Check", mock.Anything, target.URL, time.Duration(target.Timeout)*time.Second).Return(&checker.CheckResponse{
		Duration:   100 * time.Millisecond,
		Error:      nil,
		StatusCode: 404,
	}, nil)

	result, err := svc.CheckTargetSystem(ctx, targetID)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "down", result.Status)
	require.Equal(t, 404, result.StatusCode)
}

func TestCheckService_CheckTargetForUser_200(t *testing.T) {
	ctx := context.Background()
	targetID := int64(1)
	userID := int64(1)

	checkRepo := mocks.NewCheckLogRepository(t)
	targetRepo := mocks.NewTargetRepository(t)
	checkMock := mocks.NewChecker(t)

	srv := service.NewCheckService(checkRepo, targetRepo, checkMock)

	target := &models.Target{
		ID:      targetID,
		Timeout: 5,
		URL:     "https://example.com",
		UserID:  userID,
	}

	targetRepo.On("GetTargetByID", mock.Anything, targetID).Return(target, nil)
	checkMock.On("Check", mock.Anything, target.URL, time.Duration(target.Timeout)*time.Second).Return(
		&checker.CheckResponse{
			StatusCode: 200,
			Error:      nil,
			Duration:   100 * time.Millisecond,
		}, nil)
	checkRepo.On("CreateCheckLog", mock.Anything, mock.MatchedBy(func(log *models.CheckLog) bool {
		return log.StatusCode == 200 && log.ErrorMsg == nil && log.Status == "up" && log.TargetID == targetID
	})).Return(nil)

	result, err := srv.CheckTargetForUser(ctx, targetID, userID)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 200, result.StatusCode)
	require.Equal(t, "up", result.Status)
}

func TestCheckService_CheckTargetSystem_TargetNotFoundReturnsErrNoTargetFound(t *testing.T) {
	ctx := context.Background()
	targetID := int64(1)

	checkRepo := mocks.NewCheckLogRepository(t)
	targetRepo := mocks.NewTargetRepository(t)
	checkMock := mocks.NewChecker(t)

	svc := service.NewCheckService(checkRepo, targetRepo, checkMock)

	targetRepo.
		On("GetTargetByID", mock.Anything, targetID).
		Return((*models.Target)(nil), sql.ErrNoRows)

	result, err := svc.CheckTargetSystem(ctx, targetID)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, service.ErrNoTargetFound)
}

func TestCheckService_CheckTargetForUser_TargetBelongsToAnotherUserReturnsErrAccessDenied(t *testing.T) {
	ctx := context.Background()
	targetID := int64(1)
	userID := int64(1)
	ownerID := int64(2)

	checkRepo := mocks.NewCheckLogRepository(t)
	targetRepo := mocks.NewTargetRepository(t)
	checkMock := mocks.NewChecker(t)

	svc := service.NewCheckService(checkRepo, targetRepo, checkMock)

	target := &models.Target{
		ID:      targetID,
		UserID:  ownerID,
		Timeout: 5,
		URL:     "https://example.com",
	}

	targetRepo.
		On("GetTargetByID", mock.Anything, targetID).
		Return(target, nil)

	result, err := svc.CheckTargetForUser(ctx, targetID, userID)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, service.ErrAccessDenied)

	checkMock.AssertNotCalled(t, "Check", mock.Anything, mock.Anything, mock.Anything)
	checkRepo.AssertNotCalled(t, "CreateCheckLog", mock.Anything, mock.Anything)
}

func TestCheckService_CheckTargetSystem_NetworkErrorCreatesDownLogWithErrorMsg(t *testing.T) {
	ctx := context.Background()
	targetID := int64(1)

	checkRepo := mocks.NewCheckLogRepository(t)
	targetRepo := mocks.NewTargetRepository(t)
	checkMock := mocks.NewChecker(t)

	svc := service.NewCheckService(checkRepo, targetRepo, checkMock)

	target := &models.Target{
		ID:      targetID,
		Timeout: 5,
		URL:     "https://example.com",
	}

	networkErr := errors.New("connection refused")

	targetRepo.
		On("GetTargetByID", mock.Anything, targetID).
		Return(target, nil)

	checkMock.
		On("Check", mock.Anything, target.URL, time.Duration(target.Timeout)*time.Second).
		Return(&checker.CheckResponse{
			StatusCode: 0,
			Error:      networkErr,
			Duration:   150 * time.Millisecond,
		}, nil)

	checkRepo.
		On("CreateCheckLog", mock.Anything, mock.MatchedBy(func(log *models.CheckLog) bool {
			return log.TargetID == targetID &&
				log.Status == "down" &&
				log.StatusCode == 0 &&
				log.ErrorMsg != nil &&
				*log.ErrorMsg == networkErr.Error()
		})).
		Return(nil)

	result, err := svc.CheckTargetSystem(ctx, targetID)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "down", result.Status)
	require.Equal(t, 0, result.StatusCode)
	require.NotNil(t, result.ErrorMsg)
	require.Equal(t, networkErr.Error(), *result.ErrorMsg)
}

package service_test

import (
	"context"
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
		URL:     "https://example.com",
		Timeout: 5,
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

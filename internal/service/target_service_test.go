package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/thwqsz/uptime-monitor/internal/mocks"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/service"
)

func TestTargetService_CreateTarget_InvalidURLReturnsErrInvalidURL(t *testing.T) {
	ctx := context.Background()

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	result, err := svc.CreateTarget(ctx, 1, "example.com", 5, 30)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, service.ErrInvalidURL)

	repo.AssertNotCalled(t, "CreateTarget", mock.Anything, mock.Anything)
	scheduler.AssertNotCalled(t, "StartTarget", mock.Anything)
}

func TestTargetService_CreateTarget_InvalidIntervalReturnsErrInvalidInterval(t *testing.T) {
	ctx := context.Background()

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	result, err := svc.CreateTarget(ctx, 1, "https://example.com", 5, 0)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, service.ErrInvalidInterval)

	repo.AssertNotCalled(t, "CreateTarget", mock.Anything, mock.Anything)
	scheduler.AssertNotCalled(t, "StartTarget", mock.Anything)
}

func TestTargetService_CreateTarget_InvalidTimeoutReturnsErrInvalidTimeout(t *testing.T) {
	ctx := context.Background()

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	result, err := svc.CreateTarget(ctx, 1, "https://example.com", 0, 30)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, service.ErrInvalidTimeout)

	repo.AssertNotCalled(t, "CreateTarget", mock.Anything, mock.Anything)
	scheduler.AssertNotCalled(t, "StartTarget", mock.Anything)
}

func TestTargetService_CreateTarget_SuccessCreatesTargetAndStartsScheduler(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	repo.
		On("CreateTarget", mock.Anything, mock.MatchedBy(func(target *models.Target) bool {
			return target.UserID == userID &&
				target.URL == "https://example.com" &&
				target.Timeout == 5 &&
				target.IntervalTime == 30
		})).
		Return(nil)

	scheduler.
		On("StartTarget", mock.MatchedBy(func(target *models.Target) bool {
			return target.UserID == userID &&
				target.URL == "https://example.com" &&
				target.Timeout == 5 &&
				target.IntervalTime == 30
		})).
		Return()

	result, err := svc.CreateTarget(ctx, userID, "https://example.com", 5, 30)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, userID, result.UserID)
	require.Equal(t, "https://example.com", result.URL)
	require.Equal(t, 5, result.Timeout)
	require.Equal(t, 30, result.IntervalTime)
}

func TestTargetService_CreateTarget_RepoErrorReturnsError(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	repoErr := errors.New("db error")

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	repo.
		On("CreateTarget", mock.Anything, mock.Anything).
		Return(repoErr)

	result, err := svc.CreateTarget(ctx, userID, "https://example.com", 5, 30)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, repoErr)

	scheduler.AssertNotCalled(t, "StartTarget", mock.Anything)
}

func TestTargetService_DeleteTarget_InvalidUserIDReturnsErrInvalidUserID(t *testing.T) {
	ctx := context.Background()

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	err := svc.DeleteTarget(ctx, 1, 0)

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvalidUserID)

	repo.AssertNotCalled(t, "DeleteTarget", mock.Anything, mock.Anything, mock.Anything)
	scheduler.AssertNotCalled(t, "StopTarget", mock.Anything)
}

func TestTargetService_DeleteTarget_InvalidTargetIDReturnsErrInvalidTargetID(t *testing.T) {
	ctx := context.Background()

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	err := svc.DeleteTarget(ctx, 0, 1)

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvalidTargetID)

	repo.AssertNotCalled(t, "DeleteTarget", mock.Anything, mock.Anything, mock.Anything)
	scheduler.AssertNotCalled(t, "StopTarget", mock.Anything)
}

func TestTargetService_DeleteTarget_NoRowsReturnsErrNoTargetFound(t *testing.T) {
	ctx := context.Background()
	targetID := int64(1)
	userID := int64(1)

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	repo.
		On("DeleteTarget", mock.Anything, targetID, userID).
		Return(int64(0), nil)

	err := svc.DeleteTarget(ctx, targetID, userID)

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrNoTargetFound)

	scheduler.AssertNotCalled(t, "StopTarget", targetID)
}

func TestTargetService_DeleteTarget_OneRowStopsScheduler(t *testing.T) {
	ctx := context.Background()
	targetID := int64(1)
	userID := int64(1)

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	repo.
		On("DeleteTarget", mock.Anything, targetID, userID).
		Return(int64(1), nil)

	scheduler.
		On("StopTarget", targetID).
		Return()

	err := svc.DeleteTarget(ctx, targetID, userID)

	require.NoError(t, err)
}

func TestTargetService_DeleteTarget_MultipleRowsReturnsErrMultipleTargetsDeleted(t *testing.T) {
	ctx := context.Background()
	targetID := int64(1)
	userID := int64(1)

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	repo.
		On("DeleteTarget", mock.Anything, targetID, userID).
		Return(int64(2), nil)

	err := svc.DeleteTarget(ctx, targetID, userID)

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrMultipleTargetsDeleted)

	scheduler.AssertNotCalled(t, "StopTarget", targetID)
}

func TestTargetService_ListTargets_InvalidUserIDReturnsErrInvalidUserID(t *testing.T) {
	ctx := context.Background()

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	result, err := svc.ListTargets(ctx, 0)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, service.ErrInvalidUserID)

	repo.AssertNotCalled(t, "GetTargetsByUserID", mock.Anything, mock.Anything)
}

func TestTargetService_ListTargets_SuccessReturnsTargets(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)

	repo := mocks.NewTargetRepository(t)
	scheduler := mocks.NewSchedulerControl(t)

	svc := service.NewTargetService(repo, scheduler)

	targets := []*models.Target{
		{ID: 1, UserID: userID, URL: "https://a.com", Timeout: 5, IntervalTime: 30},
		{ID: 2, UserID: userID, URL: "https://b.com", Timeout: 10, IntervalTime: 60},
	}

	repo.
		On("GetTargetsByUserID", mock.Anything, userID).
		Return(targets, nil)

	result, err := svc.ListTargets(ctx, userID)

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, targets, result)
}

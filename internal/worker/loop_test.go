package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/thwqsz/uptime-monitor/internal/mocks"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"go.uber.org/zap"
)

func TestStartTarget_EnqueuesImmediateJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checker := mocks.NewTargetChecker(t)
	loop := NewLoop(nil, checker, 1, zap.NewNop(), ctx)

	target := &models.Target{
		ID:           42,
		IntervalTime: 1,
	}

	loop.StartTarget(target)

	select {
	case job := <-loop.jobs:
		require.Equal(t, int64(42), job)
	case <-time.After(300 * time.Millisecond):
		t.Fatal("expected immediate job from StartTarget")
	}

	loop.StopTarget(target.ID)
}

func TestStopTarget_StopsFurtherScheduling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checker := mocks.NewTargetChecker(t)
	loop := NewLoop(nil, checker, 1, zap.NewNop(), ctx)

	target := &models.Target{
		ID:           100,
		IntervalTime: 1,
	}

	loop.StartTarget(target)

	select {
	case <-loop.jobs:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("expected first immediate job")
	}

	loop.StopTarget(target.ID)

	select {
	case job := <-loop.jobs:
		t.Fatalf("did not expect another job after StopTarget, got %d", job)
	case <-time.After(1200 * time.Millisecond):
	}
}

func TestWorker_ProcessesJobsAndStopsWhenChannelClosed(t *testing.T) {
	checker := mocks.NewTargetChecker(t)
	loop := NewLoop(nil, checker, 1, zap.NewNop(), context.Background())

	loop.workerWg.Add(1)

	checker.
		On("CheckTargetSystem", mock.Anything, int64(1)).
		Return(&models.CheckLog{TargetID: 1}, nil).
		Once()

	checker.
		On("CheckTargetSystem", mock.Anything, int64(2)).
		Return(&models.CheckLog{TargetID: 2}, nil).
		Once()

	done := make(chan struct{})
	go func() {
		loop.worker()
		close(done)
	}()

	loop.jobs <- 1
	loop.jobs <- 2
	close(loop.jobs)

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("worker did not stop after jobs channel was closed")
	}
}

func TestRun_StartsWorkersProcessesInitialTargetsAndShutsDownGracefully(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	source := mocks.NewGetAllTargeter(t)
	checker := mocks.NewTargetChecker(t)

	targets := []*models.Target{
		{
			ID:           7,
			IntervalTime: 10,
		},
	}

	source.
		On("GetAllTargets", mock.Anything).
		Return(targets, nil).
		Once()

	called := make(chan int64, 1)

	checker.
		On("CheckTargetSystem", mock.Anything, int64(7)).
		Run(func(args mock.Arguments) {
			called <- args.Get(1).(int64)
		}).
		Return(&models.CheckLog{TargetID: 7}, nil).
		Once()

	loop := NewLoop(source, checker, 1, zap.NewNop(), ctx)

	done := make(chan struct{})
	go func() {
		loop.Run()
		close(done)
	}()

	select {
	case got := <-called:
		require.Equal(t, int64(7), got)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected initial target to be checked")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

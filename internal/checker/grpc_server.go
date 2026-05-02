package checker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thwqsz/uptime-monitor/internal/grpc/checkerpb"
)

type httpChecker interface {
	Check(ctx context.Context, url string, timeout time.Duration) (*CheckResponse, error)
}
type GRPCServerChecker struct {
	checkerService httpChecker
	checkerpb.UnimplementedCheckerServiceServer
}

func NewGRPCServerChecker(checkerService httpChecker) *GRPCServerChecker {
	return &GRPCServerChecker{checkerService: checkerService}
}

func BuildGRPCResponse(resCheck *CheckResponse) *checkerpb.CheckResponse {
	respTimeInt := int32(resCheck.Duration.Milliseconds())
	var status string
	if resCheck.StatusCode/100 == 2 && resCheck.Error == nil {
		status = "up"
	} else {
		status = "down"
	}
	var errorMsg string
	if resCheck.Error != nil {
		msg := fmt.Sprint(resCheck.Error)
		errorMsg = msg
	}
	ans := checkerpb.CheckResponse{
		StatusCode:     int32(resCheck.StatusCode),
		ResponseTimeMs: respTimeInt,
		ErrMsg:         errorMsg,
		Status:         status,
	}
	return &ans
}

func (g *GRPCServerChecker) Check(ctx context.Context, in *checkerpb.CheckRequest) (*checkerpb.CheckResponse, error) {
	if in == nil {
		return nil, errors.New("invalid input")
	}
	if in.TimeoutSec <= 0 {
		return nil, errors.New("invalid TimeOut")
	}
	if in.Url == "" {
		return nil, errors.New("invalid url")
	}
	timeDur := time.Duration(in.TimeoutSec) * time.Second
	resp, err := g.checkerService.Check(ctx, in.Url, timeDur)
	if err != nil {
		return nil, err
	}
	ans := BuildGRPCResponse(resp)
	return ans, nil
}

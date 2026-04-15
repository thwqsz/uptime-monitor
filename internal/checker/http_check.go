package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/thwqsz/uptime-monitor/internal/contracts"
)

func (c *HTTPChecker) RunConsumerLoop(ctx context.Context) {
	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{"localhost:9092"},
			Topic:       "check_tasks",
			StartOffset: kafka.FirstOffset,
		})
	defer reader.Close()
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "check_results",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()
	for {
		if ctx.Err() != nil {
			return
		}
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Println(err)
			continue
		}
		var taskStruct contracts.CheckTask
		err = json.Unmarshal(msg.Value, &taskStruct)
		if err != nil {
			log.Println(err)
			continue
		}
		res, err := c.ExecuteTask(ctx, &taskStruct)
		if err != nil {
			log.Println(err)
			continue
		}
		resByte, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			continue
		}
		err = writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte("check_result"),
			Value: resByte,
		})
		if err != nil {
			log.Println(err)
			continue
		}
	}

}

func (c *HTTPChecker) ExecuteTask(ctx context.Context, task *contracts.CheckTask) (*contracts.CheckResult, error) {
	dur := time.Duration(task.TimeoutSec) * time.Second
	resp, err := c.Check(ctx, task.URL, dur)
	if err != nil {
		return nil, err
	}
	finalResp, err := FinalFormattedCheck(resp, task)
	if err != nil {
		return nil, err
	}
	return finalResp, nil
}

type Checker interface {
	Check(ctx context.Context, url string, timeout time.Duration) (*CheckResponse, error)
}
type CheckResponse struct {
	StatusCode int
	Error      error
	Duration   time.Duration
}

type HTTPChecker struct {
	Client *http.Client
}

func NewHTTPChecker(client *http.Client) *HTTPChecker {
	return &HTTPChecker{
		Client: client,
	}
}

func (c *HTTPChecker) Check(ctx context.Context, url string, timeout time.Duration) (*CheckResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	resp, errNet := c.Client.Do(req)
	if errNet != nil {
		duration := time.Since(start)
		checkResponse := CheckResponse{
			Error:    errNet,
			Duration: duration,
		}
		return &checkResponse, nil
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	checkResponse := CheckResponse{
		StatusCode: resp.StatusCode,
		Duration:   duration,
	}
	return &checkResponse, nil

}

func FinalFormattedCheck(resCheck *CheckResponse, task *contracts.CheckTask) (*contracts.CheckResult, error) {
	respTimeInt := int(resCheck.Duration.Milliseconds())
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
	now := time.Now().UTC().Format(time.RFC3339)
	ans := contracts.CheckResult{
		TaskID:         task.TaskID,
		TargetID:       task.TargetID,
		StatusCode:     resCheck.StatusCode,
		ResponseTimeMs: respTimeInt,
		ErrorMsg:       errorMsg,
		Status:         status,
		CheckedAt:      now,
	}
	return &ans, nil
}

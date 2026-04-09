package checker

import (
	"context"
	"net/http"
	"time"
)

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
